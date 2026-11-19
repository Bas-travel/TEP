import 'package:flutter/material.dart';

import '../../core/theme.dart';

class TepBrandMark extends StatelessWidget {
  const TepBrandMark({super.key, this.compact = false, this.onSurface = false});

  final bool compact;
  final bool onSurface;

  @override
  Widget build(BuildContext context) {
    final titleColor = onSurface ? Colors.white : TepColors.ink;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: compact ? 44 : 56,
          height: compact ? 44 : 56,
          decoration: BoxDecoration(
            color: onSurface ? Colors.white : TepColors.primaryLight,
            borderRadius: BorderRadius.circular(compact ? 14 : 18),
            boxShadow: onSurface
                ? null
                : const [
                    BoxShadow(
                        color: Color(0x140C3A2A),
                        blurRadius: 14,
                        offset: Offset(0, 6))
                  ],
          ),
          child: Stack(
            alignment: Alignment.center,
            children: [
              Container(
                width: compact ? 28 : 36,
                height: compact ? 28 : 36,
                decoration: const BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: LinearGradient(
                      colors: [TepColors.highlight, TepColors.secondary]),
                ),
              ),
              Icon(Icons.flight_rounded,
                  size: compact ? 24 : 30, color: TepColors.primaryDeep),
            ],
          ),
        ),
        const SizedBox(width: 10),
        Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('TEP',
                style: TextStyle(
                    color: titleColor,
                    fontSize: compact ? 20 : 26,
                    fontWeight: FontWeight.w900,
                    letterSpacing: 1.5)),
            Text('Travel, explored together',
                style: TextStyle(
                    color: onSurface ? Colors.white70 : TepColors.muted,
                    fontSize: compact ? 9 : 11,
                    fontWeight: FontWeight.w600)),
          ],
        ),
      ],
    );
  }
}

class TepPageHeader extends StatelessWidget {
  const TepPageHeader({
    super.key,
    required this.title,
    this.subtitle,
    this.leading,
    this.actions = const [],
  });

  final String title;
  final String? subtitle;
  final Widget? leading;
  final List<Widget> actions;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          if (leading != null) ...[leading!, const SizedBox(width: 12)],
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title,
                    style: Theme.of(context)
                        .textTheme
                        .titleLarge
                        ?.copyWith(fontWeight: FontWeight.w800)),
                if (subtitle != null) ...[
                  const SizedBox(height: 2),
                  Text(subtitle!,
                      style: Theme.of(context)
                          .textTheme
                          .bodySmall
                          ?.copyWith(color: TepColors.muted)),
                ],
              ],
            ),
          ),
          ...actions,
        ],
      ),
    );
  }
}
