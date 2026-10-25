import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'listing_detail_screen.dart';
import 'widgets/common.dart';

/// Board 1.28 — city guide: pick a destination.
class CityGuideListScreen extends StatelessWidget {
  const CityGuideListScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('ไกด์เมือง')),
      body: AsyncBody<List<Destination>>(
        load: api.destinations,
        builder: (context, items, reload) => RefreshIndicator(
          onRefresh: reload,
          child: ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: items.length,
            separatorBuilder: (_, __) => const SizedBox(height: 12),
            itemBuilder: (context, i) {
              final d = items[i];
              return InkWell(
                key: Key('guide-${d.slug}'),
                borderRadius: BorderRadius.circular(18),
                onTap: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => CityGuideScreen(destination: d))),
                child: Container(
                  height: 110,
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(gradient: gradientFromCss(d.gradientCss), borderRadius: BorderRadius.circular(18)),
                  alignment: Alignment.bottomLeft,
                  child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text(d.name, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.w800)),
                    Text(d.subtitle, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.white70)),
                  ]),
                ),
              );
            },
          ),
        ),
      ),
    );
  }
}

/// A destination's guide: hotels, food and attractions grouped by type.
class CityGuideScreen extends StatelessWidget {
  const CityGuideScreen({super.key, required this.destination});
  final Destination destination;

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: Text(destination.name, maxLines: 1, overflow: TextOverflow.ellipsis)),
      body: AsyncBody<ListingPage>(
        load: () => api.listings(destinationId: destination.id, limit: 50),
        builder: (context, page, reload) {
          final byType = <String, List<ListingSummary>>{};
          for (final l in page.items) {
            byType.putIfAbsent(l.type, () => []).add(l);
          }
          return ListView(padding: EdgeInsets.zero, children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(gradient: gradientFromCss(destination.gradientCss)),
              child: Text(destination.subtitle, style: const TextStyle(color: Colors.white, fontSize: 16)),
            ),
            if (page.items.isEmpty) const SizedBox(height: 300, child: EmptyState(Icons.explore_outlined, 'ยังไม่มีข้อมูลสำหรับเมืองนี้')),
            for (final e in byType.entries) ...[
              SectionHeader(typeLabels[e.key] ?? e.key),
              for (final l in e.value)
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
                  child: ListingTile(l, onTap: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => ListingDetailScreen(listingId: l.id)))),
                ),
            ],
            const SizedBox(height: 16),
          ]);
        },
      ),
    );
  }
}
