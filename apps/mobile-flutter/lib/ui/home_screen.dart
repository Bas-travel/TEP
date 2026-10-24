import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'listing_detail_screen.dart';
import 'shell.dart';
import 'widgets/common.dart';

class _HomeData {
  const _HomeData(this.destinations, this.featured);
  final List<Destination> destinations;
  final List<ListingSummary> featured;
}

/// Board 1.1 — greeting header, search entry, categories, destinations, recommended listings.
class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key, required this.onSearch, this.modeLabel = 'mock'});
  final void Function(SearchIntent) onSearch;
  final String modeLabel;

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    final user = AppScope.of(context).user;
    return AsyncBody<_HomeData>(
      load: () async {
        final d = await api.destinations();
        final l = await api.listings(limit: 6);
        return _HomeData(d, l.items);
      },
      builder: (context, data, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(padding: EdgeInsets.zero, children: [
          Container(
            decoration: const BoxDecoration(
              color: TepColors.navy,
              borderRadius: BorderRadius.vertical(bottom: Radius.circular(28)),
            ),
            padding: EdgeInsets.fromLTRB(16, MediaQuery.of(context).padding.top + 12, 16, 20),
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Row(children: [
                CircleAvatar(
                  backgroundColor: Colors.white24,
                  child: Text((user?.firstName.isNotEmpty ?? false) ? user!.firstName.characters.first : '?', style: const TextStyle(color: Colors.white)),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text('สวัสดี, ${user?.firstName ?? ''}', style: const TextStyle(color: Colors.white, fontSize: 20, fontWeight: FontWeight.w700)),
                    const Text('Explore Your Journey', style: TextStyle(color: Colors.white70, fontSize: 13)),
                  ]),
                ),
                Chip(label: Text(modeLabel.toUpperCase(), style: const TextStyle(fontSize: 10)), visualDensity: VisualDensity.compact),
              ]),
              const SizedBox(height: 16),
              InkWell(
                key: const Key('home-search'),
                borderRadius: BorderRadius.circular(14),
                onTap: () => onSearch(const SearchIntent()),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
                  decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(14)),
                  child: const Row(children: [
                    Icon(Icons.search_rounded, color: TepColors.muted),
                    SizedBox(width: 8),
                    Text('ค้นหาที่พัก ร้านอาหาร ที่เที่ยว', style: TextStyle(color: TepColors.muted)),
                  ]),
                ),
              ),
            ]),
          ),
          SizedBox(
            height: 108,
            child: ListView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.fromLTRB(12, 14, 12, 0),
              children: [
                for (final e in typeLabels.entries)
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    child: InkWell(
                      key: Key('cat-${e.key}'),
                      borderRadius: BorderRadius.circular(14),
                      onTap: () => onSearch(SearchIntent(type: e.key)),
                      child: SizedBox(
                        width: 72,
                        child: Column(children: [
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(14), border: Border.all(color: TepColors.line)),
                            child: Icon(typeIcon(e.key), color: TepColors.navyBar),
                          ),
                          const SizedBox(height: 6),
                          Text(e.value, style: const TextStyle(fontSize: 12)),
                        ]),
                      ),
                    ),
                  ),
              ],
            ),
          ),
          const SectionHeader('จุดหมายยอดนิยม'),
          SizedBox(
            height: 150,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 16),
              itemCount: data.destinations.length,
              separatorBuilder: (_, __) => const SizedBox(width: 12),
              itemBuilder: (context, i) {
                final d = data.destinations[i];
                return InkWell(
                  key: Key('dest-${d.slug}'),
                  borderRadius: BorderRadius.circular(18),
                  onTap: () => onSearch(SearchIntent(destinationId: d.id, destinationName: d.name)),
                  child: Container(
                    width: 150,
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(gradient: gradientFromCss(d.gradientCss), borderRadius: BorderRadius.circular(18)),
                    child: Column(mainAxisAlignment: MainAxisAlignment.end, crossAxisAlignment: CrossAxisAlignment.start, children: [
                      Text(d.name, style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.w800)),
                      const SizedBox(height: 2),
                      Text(d.subtitle, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.white70, fontSize: 12)),
                    ]),
                  ),
                );
              },
            ),
          ),
          const SectionHeader('แนะนำสำหรับคุณ'),
          for (final l in data.featured)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
              child: ListingTile(l, onTap: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => ListingDetailScreen(listingId: l.id)))),
            ),
          const SizedBox(height: 16),
        ]),
      ),
    );
  }
}
