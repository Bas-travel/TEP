import 'package:flutter/material.dart';

import '../../app_scope.dart';
import '../../core/format.dart';
import '../../core/theme.dart';
import '../../data/api_client.dart';
import '../../data/models.dart';
import '../widgets/common.dart';

// ---- JSON helpers (portal pages work on the maps returned by ApiClient.portal) ----

String str(Object? v) => v == null ? '' : '$v';
num numOf(Object? v) => v is num ? v : 0;
List<Json> jlist(Object? v) => ((v as List?) ?? const []).map((e) => (e as Map).cast<String, Object?>()).toList();
Json jmap(Object? v) => v is Map ? v.cast<String, Object?>() : <String, Object?>{};
String dateOf(Object? v) {
  final d = v is String ? DateTime.tryParse(v) : null;
  return d == null ? '' : dmy(d);
}

const roleLabels = {'traveler': 'นักท่องเที่ยว', 'partner': 'พาร์ทเนอร์', 'staff': 'พนักงาน', 'moderator': 'ผู้ดูแลชุมชน', 'admin': 'แอดมิน'};
const partnerStatusLabels = {'pending': 'รออนุมัติ', 'approved': 'อนุมัติแล้ว', 'rejected': 'ไม่อนุมัติ', 'suspended': 'ระงับ'};
const listingStatusLabels = {'draft': 'ฉบับร่าง', 'published': 'เปิดขาย', 'archived': 'เก็บถาวร'};
const ticketStatusLabels = {'open': 'รอตอบ', 'answered': 'ตอบแล้ว', 'resolved': 'ปิดเรื่อง'};
const userStatusLabels = {'active': 'ใช้งาน', 'suspended': 'ถูกระงับ', 'deleted': 'ลบแล้ว'};
const verificationLabels = {'none': 'ยังไม่ตรวจ', 'verified': 'ตรวจแล้ว', 'flagged': 'ผิดปกติ'};

Color toneOf(String s) => switch (s) {
      'approved' || 'published' || 'verified' || 'answered' || 'active' || 'completed' || 'confirmed' => TepColors.success,
      'rejected' || 'suspended' || 'flagged' || 'cancelled' || 'archived' || 'deleted' => TepColors.danger,
      _ => TepColors.accent,
    };

class Pill extends StatelessWidget {
  const Pill(this.text, {super.key, this.color = TepColors.muted});
  final String text;
  final Color color;
  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
        decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(20)),
        child: Text(text, style: TextStyle(color: color, fontWeight: FontWeight.w700, fontSize: 12)),
      );
}

class StatePill extends StatelessWidget {
  const StatePill(this.state, this.labels, {super.key});
  final String state;
  final Map<String, String> labels;
  @override
  Widget build(BuildContext context) => Pill(labels[state] ?? state, color: toneOf(state));
}

// ---- dashboards ----

class Kpi {
  const Kpi(this.label, this.value, this.icon);
  final String label, value;
  final IconData icon;
}

class KpiGrid extends StatelessWidget {
  const KpiGrid(this.items, {super.key});
  final List<Kpi> items;
  @override
  Widget build(BuildContext context) => GridView.count(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        crossAxisCount: 2,
        mainAxisSpacing: 10,
        crossAxisSpacing: 10,
        childAspectRatio: 1.6,
        children: [
          for (final k in items)
            TepCard(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, mainAxisAlignment: MainAxisAlignment.center, children: [
                Icon(k.icon, color: TepColors.primary),
                const SizedBox(height: 4),
                Text(k.value, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800)),
                Text(k.label, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 12)),
              ]),
            ),
        ],
      );
}

/// Loads one portal endpoint and renders it with pull-to-refresh. Used by the dashboards and report pages.
class PortalPage extends StatelessWidget {
  const PortalPage({super.key, required this.path, required this.builder, this.query = const {}});
  final String path;
  final Map<String, String?> query;
  final List<Widget> Function(BuildContext context, Json data) builder;
  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return AsyncBody<Json>(
      load: () => api.portal('GET', path, query: query),
      builder: (context, data, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(physics: const AlwaysScrollableScrollPhysics(), padding: const EdgeInsets.all(16), children: builder(context, data)),
      ),
    );
  }
}

// ---- dialogs & actions ----

/// Runs a mutating portal call, shows an error SnackBar (ApiException) or the optional success message. True when it succeeded.
Future<bool> act(BuildContext context, Future<Json> Function() call, {String? ok}) async {
  final r = await guarded<Json>(context, call);
  if (r == null) return false;
  if (ok != null && context.mounted) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(ok)));
  }
  return true;
}

Future<bool> confirmDialog(BuildContext context, String title, String message) async {
  final r = await showDialog<bool>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: Text(title),
      content: Text(message),
      actions: [
        TextButton(onPressed: () => Navigator.of(ctx).pop(false), child: const Text('ยกเลิก')),
        TextButton(key: const Key('dialog-confirm'), onPressed: () => Navigator.of(ctx).pop(true), child: const Text('ยืนยัน')),
      ],
    ),
  );
  return r ?? false;
}

/// Asks for a short text (reason / note). Returns null when cancelled.
Future<String?> askText(BuildContext context, {required String title, required String label, bool mandatory = true, int maxLength = 450, String initial = '', TextInputType? keyboard}) =>
    showDialog<String>(
      context: context,
      builder: (ctx) => _TextPrompt(title: title, label: label, mandatory: mandatory, maxLength: maxLength, initial: initial, keyboard: keyboard),
    );

class _TextPrompt extends StatefulWidget {
  const _TextPrompt({required this.title, required this.label, required this.mandatory, required this.maxLength, required this.initial, this.keyboard});
  final String title, label, initial;
  final bool mandatory;
  final int maxLength;
  final TextInputType? keyboard;
  @override
  State<_TextPrompt> createState() => _TextPromptState();
}

class _TextPromptState extends State<_TextPrompt> {
  late final _c = TextEditingController(text: widget.initial);
  @override
  void dispose() {
    _c.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
        title: Text(widget.title),
        content: TextField(
          key: const Key('prompt-text'),
          controller: _c,
          autofocus: true,
          maxLines: widget.keyboard == null ? 3 : 1,
          maxLength: widget.maxLength,
          keyboardType: widget.keyboard,
          decoration: InputDecoration(labelText: widget.label),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.of(context).pop(), child: const Text('ยกเลิก')),
          TextButton(
            key: const Key('prompt-ok'),
            onPressed: () {
              final t = _c.text.trim();
              if (widget.mandatory && t.isEmpty) return;
              Navigator.of(context).pop(t);
            },
            child: const Text('ตกลง'),
          ),
        ],
      );
}

// ---- paged list with search + filter chips ----

class ListFilter {
  const ListFilter(this.label, [this.query = const {}]);
  final String label;
  final Map<String, String> query;
}

class PortalList extends StatefulWidget {
  const PortalList({
    super.key,
    required this.path,
    required this.itemBuilder,
    this.filters = const [],
    this.searchHint,
    this.emptyText = 'ไม่พบข้อมูล',
    this.pageSize = 20,
  });
  final String path;
  final Widget Function(BuildContext context, Json item, Future<void> Function() reload) itemBuilder;
  final List<ListFilter> filters;
  final String? searchHint;
  final String emptyText;
  final int pageSize;
  @override
  State<PortalList> createState() => _PortalListState();
}

class _PortalListState extends State<PortalList> {
  final _search = TextEditingController();
  List<Json> _items = [];
  int? _total;
  int _filter = 0;
  bool _loading = true, _more = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  Map<String, String?> _query(int offset) => {
        if (widget.filters.isNotEmpty) ...widget.filters[_filter].query,
        if (_search.text.trim().isNotEmpty) 'q': _search.text.trim(),
        'offset': '$offset',
        'limit': '${widget.pageSize}',
      };

  Future<void> _load({bool reset = true}) async {
    final api = AppScope.apiOf(context);
    setState(() {
      if (reset) {
        _loading = true;
        _error = null;
      } else {
        _more = true;
      }
    });
    try {
      final j = await api.portal('GET', widget.path, query: _query(reset ? 0 : _items.length));
      if (!mounted) return;
      final got = jlist(j['items']);
      setState(() {
        _items = reset ? got : [..._items, ...got];
        _total = j['total'] is num ? (j['total'] as num).toInt() : null;
        _loading = false;
        _more = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.userMessage;
        _loading = false;
        _more = false;
      });
    }
  }

  Future<void> _reload() => _load();

  @override
  Widget build(BuildContext context) {
    final hasMore = _total != null && _items.length < _total!;
    return Column(children: [
      if (widget.searchHint != null)
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
          child: TextField(
            key: const Key('portal-search'),
            controller: _search,
            textInputAction: TextInputAction.search,
            onSubmitted: (_) => _load(),
            decoration: InputDecoration(
              hintText: widget.searchHint,
              prefixIcon: const Icon(Icons.search_rounded),
              suffixIcon: IconButton(
                icon: const Icon(Icons.clear_rounded),
                onPressed: () {
                  _search.clear();
                  _load();
                },
              ),
              isDense: true,
            ),
          ),
        ),
      if (widget.filters.length > 1)
        SizedBox(
          height: 52,
          child: ListView(scrollDirection: Axis.horizontal, padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8), children: [
            for (var i = 0; i < widget.filters.length; i++)
              Padding(
                padding: const EdgeInsets.only(right: 8),
                child: ChoiceChip(
                  key: Key('filter-${widget.filters[i].label}'),
                  label: Text(widget.filters[i].label),
                  selected: _filter == i,
                  onSelected: (_) {
                    setState(() => _filter = i);
                    _load();
                  },
                ),
              ),
          ]),
        ),
      Expanded(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
                ? Center(
                    child: Column(mainAxisSize: MainAxisSize.min, children: [
                      Text(_error!, textAlign: TextAlign.center),
                      const SizedBox(height: 12),
                      OutlinedButton(onPressed: _reload, child: const Text('ลองอีกครั้ง')),
                    ]),
                  )
                : RefreshIndicator(
                    onRefresh: _reload,
                    child: _items.isEmpty
                        ? ListView(children: [SizedBox(height: 260, child: EmptyState(Icons.inbox_outlined, widget.emptyText))])
                        : ListView(
                            physics: const AlwaysScrollableScrollPhysics(),
                            padding: const EdgeInsets.fromLTRB(16, 8, 16, 88),
                            children: [
                              if (_total != null) Padding(padding: const EdgeInsets.only(bottom: 8), child: Text('ทั้งหมด $_total รายการ', style: const TextStyle(color: TepColors.muted, fontSize: 12))),
                              for (final it in _items) Padding(padding: const EdgeInsets.only(bottom: 10), child: widget.itemBuilder(context, it, _reload)),
                              if (hasMore)
                                Center(
                                  child: OutlinedButton(
                                    key: const Key('load-more'),
                                    onPressed: _more ? null : () => _load(reset: false),
                                    child: Text(_more ? 'กำลังโหลด…' : 'โหลดเพิ่ม'),
                                  ),
                                ),
                            ],
                          ),
                  ),
      ),
    ]);
  }
}

/// Small compact button used inside list tiles (the theme's FilledButton is full-width by default).
Widget tileButton(String label, VoidCallback? onPressed, {Key? key, bool danger = false, bool outlined = false}) {
  final color = danger ? TepColors.danger : TepColors.primary;
  if (outlined) {
    return OutlinedButton(key: key, onPressed: onPressed, style: OutlinedButton.styleFrom(minimumSize: const Size(0, 38), foregroundColor: color, side: BorderSide(color: color)), child: Text(label));
  }
  return FilledButton(key: key, onPressed: onPressed, style: FilledButton.styleFrom(minimumSize: const Size(0, 38), backgroundColor: color), child: Text(label));
}
