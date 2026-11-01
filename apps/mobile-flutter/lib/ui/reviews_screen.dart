import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import '../data/models_engage.dart';
import 'widgets/common.dart';

/// Board 1.15 — my reviews, with entry to write a new one.
class ReviewsScreen extends StatefulWidget {
  const ReviewsScreen({super.key});
  @override
  State<ReviewsScreen> createState() => _ReviewsScreenState();
}

class _ReviewsScreenState extends State<ReviewsScreen> {
  int _version = 0;

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('รีวิวของฉัน')),
      floatingActionButton: FloatingActionButton.extended(
        key: const Key('new-review'),
        backgroundColor: TepColors.primary,
        foregroundColor: Colors.white,
        icon: const Icon(Icons.rate_review_rounded),
        label: const Text('เขียนรีวิว'),
        onPressed: () async {
          await Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const WriteReviewScreen()));
          if (mounted) setState(() => _version++);
        },
      ),
      body: AsyncBody<List<Review>>(
        key: ValueKey('r-$_version'),
        load: api.myReviews,
        builder: (context, items, reload) {
          if (items.isEmpty) return const EmptyState(Icons.star_border_rounded, 'ยังไม่มีรีวิว\nแชร์ประสบการณ์ให้นักเดินทางคนอื่น');
          return RefreshIndicator(
            onRefresh: reload,
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 88),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (context, i) {
                final r = items[i];
                return TepCard(
                  child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Row(children: [
                      Expanded(child: Text(r.itemName, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w800))),
                      Text('★' * r.rating, style: const TextStyle(color: TepColors.accent)),
                    ]),
                    const SizedBox(height: 4),
                    Text(r.comment),
                    const SizedBox(height: 4),
                    Text(dmy(r.createdAt), style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                  ]),
                );
              },
            ),
          );
        },
      ),
    );
  }
}

/// Write a review. With [listingId] set the item is fixed (opened from a listing); otherwise the user picks one.
class WriteReviewScreen extends StatefulWidget {
  const WriteReviewScreen({super.key, this.listingId, this.listingName});
  final int? listingId;
  final String? listingName;
  @override
  State<WriteReviewScreen> createState() => _WriteReviewScreenState();
}

class _WriteReviewScreenState extends State<WriteReviewScreen> {
  final _comment = TextEditingController();
  int _rating = 5;
  int? _listingId;
  bool _busy = false;
  late final Future<ListingPage> _options = AppScope.apiOf(context).listings(limit: 50);

  @override
  void initState() {
    super.initState();
    _listingId = widget.listingId;
  }

  @override
  void dispose() {
    _comment.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final id = _listingId;
    if (id == null) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('กรุณาเลือกสถานที่ที่ต้องการรีวิว')));
      return;
    }
    final api = AppScope.apiOf(context);
    setState(() => _busy = true);
    final r = await guarded(context, () => api.createReview(listingId: id, rating: _rating, comment: _comment.text.trim()));
    if (!mounted) return;
    setState(() => _busy = false);
    if (r == null) return;
    ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('ขอบคุณสำหรับรีวิว')));
    Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('เขียนรีวิว')),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        if (widget.listingId != null)
          TepCard(child: Row(children: [const Icon(Icons.place_rounded, color: TepColors.primary), const SizedBox(width: 8), Expanded(child: Text(widget.listingName ?? 'รายการ #${widget.listingId}', style: const TextStyle(fontWeight: FontWeight.w700)))]))
        else
          FutureBuilder<ListingPage>(
            future: _options,
            builder: (context, snap) {
              if (!snap.hasData) return const Padding(padding: EdgeInsets.all(16), child: Center(child: CircularProgressIndicator()));
              return DropdownButtonFormField<int>(
                key: const Key('review-listing'),
                isExpanded: true,
                initialValue: _listingId,
                decoration: const InputDecoration(labelText: 'เลือกสถานที่'),
                items: [for (final l in snap.data!.items) DropdownMenuItem(value: l.id, child: Text(l.name, overflow: TextOverflow.ellipsis))],
                onChanged: (v) => setState(() => _listingId = v),
              );
            },
          ),
        const SizedBox(height: 16),
        Row(mainAxisAlignment: MainAxisAlignment.center, children: [
          for (var i = 1; i <= 5; i++)
            IconButton(
              key: Key('star-$i'),
              iconSize: 36,
              onPressed: () => setState(() => _rating = i),
              icon: Icon(i <= _rating ? Icons.star_rounded : Icons.star_border_rounded, color: TepColors.accent),
            ),
        ]),
        const SizedBox(height: 8),
        TextField(key: const Key('review-comment'), controller: _comment, maxLines: 5, maxLength: 2000, decoration: const InputDecoration(labelText: 'ความคิดเห็นของคุณ')),
        const SizedBox(height: 12),
        FilledButton(key: const Key('review-submit'), onPressed: _busy ? null : _submit, child: const Text('ส่งรีวิว')),
      ]),
    );
  }
}
