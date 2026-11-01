import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'booking_flow_screen.dart';
import 'widgets/common.dart';

/// Board 1.4 / 1.13 — listing detail with sticky "จองเลย" bar.
class ListingDetailScreen extends StatelessWidget {
  const ListingDetailScreen({super.key, required this.listingId});
  final int listingId;

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('รายละเอียด')),
      body: AsyncBody<ListingDetail>(
        load: () => api.listing(listingId),
        builder: (context, l, _) => Column(children: [
          Expanded(
            child: ListView(padding: EdgeInsets.zero, children: [
              Container(
                height: 200,
                decoration: BoxDecoration(
                  gradient: gradientFromCss(l.imageUrl == null ? 'linear-gradient(135deg,#002765,#3988A2)' : ''),
                  image: (l.imageUrl != null && l.imageUrl!.isNotEmpty) ? DecorationImage(image: NetworkImage(l.imageUrl!), fit: BoxFit.cover) : null,
                ),
                alignment: Alignment.center,
                child: Icon(typeIcon(l.type), size: 72, color: Colors.white54),
              ),
              Padding(
                padding: const EdgeInsets.all(16),
                child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                  Text(typeLabels[l.type] ?? l.type, style: const TextStyle(color: TepColors.partner, fontWeight: FontWeight.w700)),
                  const SizedBox(height: 4),
                  Text(l.name, style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w800, color: TepColors.ink)),
                  const SizedBox(height: 6),
                  Row(children: [
                    const Icon(Icons.place_outlined, size: 16, color: TepColors.muted),
                    const SizedBox(width: 4),
                    Expanded(child: Text(l.location, style: const TextStyle(color: TepColors.muted))),
                  ]),
                  const SizedBox(height: 8),
                  RatingChip(l.rating, l.reviewCount),
                  const SectionHeaderInline('รายละเอียด'),
                  Text(l.description, style: const TextStyle(height: 1.5)),
                  const SectionHeaderInline('สิ่งอำนวยความสะดวก'),
                  Wrap(spacing: 8, runSpacing: 8, children: [
                    for (final a in l.amenities) Chip(label: Text(a), backgroundColor: Colors.white, side: const BorderSide(color: TepColors.line)),
                  ]),
                ]),
              ),
            ]),
          ),
          SafeArea(
            top: false,
            child: Container(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
              decoration: const BoxDecoration(color: Colors.white, border: Border(top: BorderSide(color: TepColors.line))),
              child: Row(children: [
                Column(crossAxisAlignment: CrossAxisAlignment.start, mainAxisSize: MainAxisSize.min, children: [
                  const Text('เริ่มต้น', style: TextStyle(color: TepColors.muted, fontSize: 12)),
                  Text(thb(l.priceFrom), style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w800, color: TepColors.primary)),
                ]),
                const SizedBox(width: 20),
                Expanded(
                  child: FilledButton(
                    key: const Key('book-now'),
                    onPressed: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => BookingFlowScreen(listing: l))),
                    child: const Text('จองเลย'),
                  ),
                ),
              ]),
            ),
          ),
        ]),
      ),
    );
  }
}

class SectionHeaderInline extends StatelessWidget {
  const SectionHeaderInline(this.text, {super.key});
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(top: 20, bottom: 8),
        child: Text(text, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
      );
}
