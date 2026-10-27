import 'package:flutter/material.dart';

import '../../core/format.dart';
import '../../core/theme.dart';
import '../../data/models.dart';

/// Run an async UI action and surface ApiException as a SnackBar.
Future<T?> guarded<T>(BuildContext context, Future<T> Function() action) async {
  try {
    return await action();
  } on ApiException catch (e) {
    if (context.mounted) {
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(SnackBar(content: Text(e.userMessage), backgroundColor: TepColors.danger));
    }
    return null;
  }
}

class TepCard extends StatelessWidget {
  const TepCard({super.key, required this.child, this.padding = const EdgeInsets.all(14), this.onTap});
  final Widget child;
  final EdgeInsets padding;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: TepColors.card,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16), side: const BorderSide(color: TepColors.line)),
      clipBehavior: Clip.antiAlias,
      child: InkWell(onTap: onTap, child: Padding(padding: padding, child: child)),
    );
  }
}

class SectionHeader extends StatelessWidget {
  const SectionHeader(this.title, {super.key, this.trailing});
  final String title;
  final Widget? trailing;
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.fromLTRB(16, 20, 16, 10),
        child: Row(children: [
          Expanded(child: Text(title, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: TepColors.ink))),
          if (trailing != null) trailing!,
        ]),
      );
}

class RatingChip extends StatelessWidget {
  const RatingChip(this.rating, this.count, {super.key});
  final double rating;
  final int count;
  @override
  Widget build(BuildContext context) => Row(mainAxisSize: MainAxisSize.min, children: [
        const Icon(Icons.star_rounded, size: 16, color: TepColors.accent),
        const SizedBox(width: 2),
        Text(rating.toStringAsFixed(1), style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 13)),
        Text(' ($count)', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
      ]);
}

class ListingThumb extends StatelessWidget {
  const ListingThumb({super.key, required this.type, this.imageUrl, this.size = 76, this.seed = 0});
  final String type;
  final String? imageUrl;
  final double size;
  final int seed;

  static const _palettes = [
    ['#0b5ed7', '#2bb3c0'],
    ['#1b7f5c', '#8bc34a'],
    ['#0f766e', '#38bdf8'],
    ['#002765', '#6366f1'],
    ['#f59e0b', '#ef4444'],
    ['#7c3aed', '#22c55e'],
  ];

  @override
  Widget build(BuildContext context) {
    final p = _palettes[seed % _palettes.length];
    final fallback = Container(
      width: size,
      height: size,
      decoration: BoxDecoration(gradient: gradientFromCss('${p[0]},${p[1]}'), borderRadius: BorderRadius.circular(14)),
      child: Icon(typeIcon(type), color: Colors.white, size: size * 0.4),
    );
    if (imageUrl == null || imageUrl!.isEmpty) return fallback;
    return ClipRRect(
      borderRadius: BorderRadius.circular(14),
      child: Image.network(imageUrl!, width: size, height: size, fit: BoxFit.cover, errorBuilder: (_, __, ___) => fallback),
    );
  }
}

class ListingTile extends StatelessWidget {
  const ListingTile(this.l, {super.key, required this.onTap});
  final ListingSummary l;
  final VoidCallback onTap;
  @override
  Widget build(BuildContext context) => TepCard(
        onTap: onTap,
        child: Row(children: [
          ListingThumb(type: l.type, imageUrl: l.imageUrl, seed: l.id),
          const SizedBox(width: 12),
          Expanded(
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(l.name, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15)),
              const SizedBox(height: 2),
              Text(l.location, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 13)),
              const SizedBox(height: 6),
              Row(children: [
                Expanded(child: FittedBox(fit: BoxFit.scaleDown, alignment: Alignment.centerLeft, child: RatingChip(l.rating, l.reviewCount))),
                const SizedBox(width: 8),
                Text('เริ่ม ${thb(l.priceFrom)}', style: const TextStyle(color: TepColors.primary, fontWeight: FontWeight.w700)),
              ]),
            ]),
          ),
        ]),
      );
}

class StatusBadge extends StatelessWidget {
  const StatusBadge(this.status, {super.key});
  final String status;
  @override
  Widget build(BuildContext context) {
    final color = switch (status) {
      'confirmed' || 'completed' => TepColors.success,
      'cancelled' => TepColors.danger,
      _ => TepColors.accent,
    };
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(20)),
      child: Text(statusLabels[status] ?? status, style: TextStyle(color: color, fontWeight: FontWeight.w700, fontSize: 12)),
    );
  }
}

/// Generic async body: loading / error + retry / data.
class AsyncBody<T> extends StatefulWidget {
  const AsyncBody({super.key, required this.load, required this.builder});
  final Future<T> Function() load;
  final Widget Function(BuildContext context, T data, Future<void> Function() reload) builder;
  @override
  State<AsyncBody<T>> createState() => _AsyncBodyState<T>();
}

class _AsyncBodyState<T> extends State<AsyncBody<T>> {
  late Future<T> _f = widget.load();

  Future<void> _reload() async {
    final f = widget.load();
    setState(() {
      _f = f;
    });
    try {
      await f;
    } catch (_) {
      // Error is rendered by the FutureBuilder.
    }
  }

  @override
  Widget build(BuildContext context) => FutureBuilder<T>(
        future: _f,
        builder: (context, snap) {
          if (snap.connectionState != ConnectionState.done) return const Center(child: CircularProgressIndicator());
          if (snap.hasError) {
            final e = snap.error;
            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(mainAxisSize: MainAxisSize.min, children: [
                  const Icon(Icons.cloud_off_rounded, size: 48, color: TepColors.muted),
                  const SizedBox(height: 12),
                  Text(e is ApiException ? e.userMessage : 'เกิดข้อผิดพลาด กรุณาลองใหม่', textAlign: TextAlign.center),
                  const SizedBox(height: 12),
                  OutlinedButton(onPressed: _reload, child: const Text('ลองอีกครั้ง')),
                ]),
              ),
            );
          }
          return widget.builder(context, snap.data as T, _reload);
        },
      );
}

class EmptyState extends StatelessWidget {
  const EmptyState(this.icon, this.message, {super.key});
  final IconData icon;
  final String message;
  @override
  Widget build(BuildContext context) => Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            Icon(icon, size: 56, color: TepColors.line),
            const SizedBox(height: 12),
            Text(message, textAlign: TextAlign.center, style: const TextStyle(color: TepColors.muted)),
          ]),
        ),
      );
}
