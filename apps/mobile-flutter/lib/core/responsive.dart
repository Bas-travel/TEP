import 'package:flutter/widgets.dart';

class TepBreakpoints {
  const TepBreakpoints._();

  static const compact = 600.0;
  static const medium = 840.0;
  static const expanded = 1200.0;

  static bool isCompact(BuildContext context) =>
      MediaQuery.sizeOf(context).width < compact;
  static bool isMedium(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    return width >= compact && width < medium;
  }

  static bool isExpanded(BuildContext context) =>
      MediaQuery.sizeOf(context).width >= medium;

  static int gridColumns(BuildContext context,
      {int compact = 1, int medium = 2, int expanded = 3}) {
    final width = MediaQuery.sizeOf(context).width;
    if (width >= TepBreakpoints.medium) return expanded;
    if (width >= TepBreakpoints.compact) return medium;
    return compact;
  }

  static EdgeInsets pagePadding(BuildContext context) {
    if (isExpanded(context)) return const EdgeInsets.symmetric(horizontal: 32);
    if (isMedium(context)) return const EdgeInsets.symmetric(horizontal: 24);
    return const EdgeInsets.symmetric(horizontal: 16);
  }
}

class TepConstrainedContent extends StatelessWidget {
  const TepConstrainedContent(
      {super.key, required this.child, this.maxWidth = 1100});

  final Widget child;
  final double maxWidth;

  @override
  Widget build(BuildContext context) => Center(
        child: ConstrainedBox(
          constraints: BoxConstraints(maxWidth: maxWidth),
          child: child,
        ),
      );
}
