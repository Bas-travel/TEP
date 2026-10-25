import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models_engage.dart';
import 'widgets/common.dart';

const _postTypeLabels = {'story': 'เรื่องเล่า', 'tip': 'เคล็ดลับ', 'question': 'คำถาม', 'trip': 'แชร์ทริป'};

/// Board 1.25 — community feed with all / following filter and per-destination filter.
class CommunityScreen extends StatefulWidget {
  const CommunityScreen({super.key});
  @override
  State<CommunityScreen> createState() => _CommunityScreenState();
}

class _CommunityScreenState extends State<CommunityScreen> {
  String _mode = 'all';
  String? _dest;
  final Map<int, LikeResult> _likes = {};

  Future<void> _like(Post p) async {
    final api = AppScope.apiOf(context);
    final r = await guarded(context, () => api.toggleLike(p.id));
    if (r != null && mounted) setState(() => _likes[p.id] = r);
  }

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('ชุมชนนักเดินทาง')),
      body: Column(children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
          child: Row(children: [
            ChoiceChip(key: const Key('feed-all'), label: const Text('ทั้งหมด'), selected: _mode == 'all', onSelected: (_) => setState(() => _mode = 'all')),
            const SizedBox(width: 8),
            ChoiceChip(key: const Key('feed-following'), label: const Text('ที่ติดตาม'), selected: _mode == 'following', onSelected: (_) => setState(() => _mode = 'following')),
            const SizedBox(width: 8),
            if (_dest != null)
              Flexible(
                child: InputChip(
                  key: const Key('feed-dest-clear'),
                  label: Text(_dest!, overflow: TextOverflow.ellipsis),
                  onDeleted: () => setState(() => _dest = null),
                ),
              ),
          ]),
        ),
        Expanded(
          child: AsyncBody<List<Post>>(
            key: ValueKey('feed-$_mode-$_dest'),
            load: () => api.feed(mode: _mode, destination: _dest),
            builder: (context, items, reload) {
              if (items.isEmpty) return const EmptyState(Icons.forum_outlined, 'ยังไม่มีโพสต์');
              return RefreshIndicator(
                onRefresh: reload,
                child: ListView.separated(
                  padding: const EdgeInsets.all(16),
                  itemCount: items.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 10),
                  itemBuilder: (context, i) {
                    final p = items[i];
                    final liked = _likes[p.id]?.liked ?? p.liked;
                    final count = _likes[p.id]?.likeCount ?? p.likeCount;
                    return TepCard(
                      key: Key('post-${p.id}'),
                      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                        Row(children: [
                          CircleAvatar(radius: 16, backgroundColor: TepColors.navyBar, child: Text(p.authorId.replaceFirst('m_', '').characters.firstOrNull?.toUpperCase() ?? '?', style: const TextStyle(color: Colors.white, fontSize: 12))),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                              Text(p.authorId.replaceFirst('m_', ''), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                              Text('${_postTypeLabels[p.type] ?? p.type} · ${dmy(p.createdAt)}', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                            ]),
                          ),
                        ]),
                        const SizedBox(height: 10),
                        Text(p.text),
                        if (p.hasImage) ...[
                          const SizedBox(height: 10),
                          Container(
                            height: 130,
                            decoration: BoxDecoration(gradient: gradientFromCss('#0b5ed7,#2bb3c0'), borderRadius: BorderRadius.circular(12)),
                            alignment: Alignment.center,
                            child: const Icon(Icons.image_rounded, color: Colors.white70, size: 40),
                          ),
                        ],
                        const SizedBox(height: 8),
                        Row(children: [
                          Flexible(
                            child: ActionChip(
                              key: Key('post-dest-${p.id}'),
                              avatar: const Icon(Icons.place_outlined, size: 16),
                              label: Text(p.destination, overflow: TextOverflow.ellipsis),
                              onPressed: () => setState(() => _dest = p.destination),
                            ),
                          ),
                          if (p.sharedDays != null) ...[
                            const SizedBox(width: 6),
                            Chip(label: Text('แผน ${p.sharedDays} วัน'), visualDensity: VisualDensity.compact),
                          ],
                          const Spacer(),
                          IconButton(
                            key: Key('like-${p.id}'),
                            onPressed: () => _like(p),
                            icon: Icon(liked ? Icons.favorite_rounded : Icons.favorite_border_rounded, color: liked ? TepColors.danger : TepColors.muted),
                          ),
                          Text('$count', key: Key('like-count-${p.id}')),
                        ]),
                      ]),
                    );
                  },
                ),
              );
            },
          ),
        ),
      ]),
    );
  }
}
