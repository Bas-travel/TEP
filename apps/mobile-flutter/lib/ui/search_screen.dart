import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'listing_detail_screen.dart';
import 'shell.dart';
import 'widgets/common.dart';

/// Board 1.2 + 1.3 — search box, type filter, paginated results.
class SearchScreen extends StatefulWidget {
  const SearchScreen({super.key, this.intent = const SearchIntent()});
  final SearchIntent intent;
  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  late final TextEditingController _q = TextEditingController(text: widget.intent.query ?? '');
  late String? _type = widget.intent.type;
  late int? _destId = widget.intent.destinationId;
  late String? _destName = widget.intent.destinationName;
  final _items = <ListingSummary>[];
  String? _cursor;
  bool _loading = false, _loaded = false;
  Object? _error;
  int _gen = 0; // discards stale responses

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _run(reset: true));
  }

  @override
  void dispose() {
    _q.dispose();
    super.dispose();
  }

  Future<void> _run({required bool reset}) async {
    final api = AppScope.apiOf(context);
    final gen = ++_gen;
    setState(() {
      _loading = true;
      _error = null;
      if (reset) {
        _items.clear();
        _cursor = null;
      }
    });
    try {
      final page = await api.listings(q: _q.text, type: _type, destinationId: _destId, cursor: reset ? null : _cursor);
      if (!mounted || gen != _gen) return;
      setState(() {
        _items.addAll(page.items);
        _cursor = page.nextCursor;
        _loaded = true;
      });
    } catch (e) {
      if (!mounted || gen != _gen) return;
      setState(() => _error = e);
    } finally {
      if (mounted && gen == _gen) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(_destName == null ? 'ค้นหา' : 'ค้นหา · $_destName')),
      body: Column(children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
          child: TextField(
            key: const Key('search-field'),
            controller: _q,
            textInputAction: TextInputAction.search,
            onSubmitted: (_) => _run(reset: true),
            decoration: InputDecoration(
              hintText: 'ชื่อสถานที่ / พื้นที่',
              prefixIcon: const Icon(Icons.search_rounded),
              suffixIcon: _q.text.isEmpty
                  ? null
                  : IconButton(
                      icon: const Icon(Icons.close_rounded),
                      onPressed: () {
                        _q.clear();
                        _run(reset: true);
                      },
                    ),
            ),
            onChanged: (_) => setState(() {}),
          ),
        ),
        SizedBox(
          height: 52,
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            children: [
              _chip('ทั้งหมด', _type == null && _destId == null, () {
                _type = null;
                _destId = null;
                _destName = null;
                _run(reset: true);
              }),
              for (final e in typeLabels.entries)
                _chip(e.value, _type == e.key, () {
                  _type = _type == e.key ? null : e.key;
                  _run(reset: true);
                }, key: Key('chip-${e.key}')),
            ],
          ),
        ),
        Expanded(child: _body()),
      ]),
    );
  }

  Widget _chip(String label, bool selected, VoidCallback onTap, {Key? key}) => Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 8),
        child: ChoiceChip(key: key, label: Text(label), selected: selected, onSelected: (_) => onTap(), selectedColor: TepColors.primary.withValues(alpha: 0.14)),
      );

  Widget _body() {
    if (_error != null && _items.isEmpty) {
      final e = _error;
      return Center(
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          Text(e is ApiException ? e.userMessage : 'เกิดข้อผิดพลาด'),
          const SizedBox(height: 8),
          OutlinedButton(onPressed: () => _run(reset: true), child: const Text('ลองอีกครั้ง')),
        ]),
      );
    }
    if (_loading && _items.isEmpty) return const Center(child: CircularProgressIndicator());
    if (_loaded && _items.isEmpty) return const EmptyState(Icons.search_off_rounded, 'ไม่พบรายการที่ตรงกับการค้นหา');
    return ListView.separated(
      key: const Key('results'),
      padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
      itemCount: _items.length + (_cursor != null ? 1 : 0),
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (context, i) {
        if (i == _items.length) {
          return Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: OutlinedButton(onPressed: _loading ? null : () => _run(reset: false), child: Text(_loading ? 'กำลังโหลด…' : 'โหลดเพิ่ม')),
          );
        }
        final l = _items[i];
        return ListingTile(l, onTap: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => ListingDetailScreen(listingId: l.id))));
      },
    );
  }
}
