import 'package:flutter/material.dart';

/// TEP brand tokens — "Sunset Journey".
/// Brand board: warm, friendly, playful, and WCAG-conscious.
class TepColors {
  static const primary = Color(0xFFC2410C);
  static const primaryDeep = Color(0xFF9A3412);
  static const primaryLight = Color(0xFFFFF0E8);
  static const secondary = Color(0xFFFB923C);
  static const coral = Color(0xFFE85D75);
  static const highlight = Color(0xFFEAB308);
  static const ink = Color(0xFF18392B);
  static const muted = Color(0xFF647067);
  static const background = Color(0xFFF3FBF5);
  static const card = Colors.white;
  static const line = Color(0xFFE2E8E3);
  static const success = Color(0xFF15803D);
  static const warning = Color(0xFFB45309);
  static const danger = Color(0xFFB91C1C);
  static const info = Color(0xFF1D4ED8);

  // Compatibility token aliases.
  static const accent = highlight;
  static const partner = Color(0xFF3988A2);
  static const admin = Color(0xFF7C7FC9);
  static const staff = Color(0xFFF59E0B);
}

class TepTheme {
  static const _fallback = ['Noto Sans Thai', 'Sarabun', 'Roboto'];

  static ThemeData light() {
    final scheme = ColorScheme.fromSeed(
      seedColor: TepColors.primary,
      brightness: Brightness.light,
    ).copyWith(
      primary: TepColors.primary,
      onPrimary: Colors.white,
      secondary: TepColors.secondary,
      onSecondary: TepColors.ink,
      surface: TepColors.card,
      onSurface: TepColors.ink,
      error: TepColors.danger,
    );
    return ThemeData(
      useMaterial3: true,
      colorScheme: scheme,
      scaffoldBackgroundColor: TepColors.background,
      fontFamilyFallback: _fallback,
      textTheme: ThemeData.light()
          .textTheme
          .apply(bodyColor: TepColors.ink, displayColor: TepColors.ink),
      appBarTheme: const AppBarTheme(
        backgroundColor: TepColors.primary,
        foregroundColor: Colors.white,
        elevation: 0,
        centerTitle: true,
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: TepColors.primary,
          foregroundColor: Colors.white,
          minimumSize: const Size.fromHeight(52),
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          textStyle: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: Colors.white,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: TepColors.line),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: TepColors.line),
        ),
      ),
      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: Colors.white,
        indicatorColor: TepColors.secondary.withValues(alpha: 0.30),
      ),
      chipTheme: ChipThemeData(
        selectedColor: TepColors.secondary.withValues(alpha: 0.35),
        side: const BorderSide(color: TepColors.line),
      ),
    );
  }
}
