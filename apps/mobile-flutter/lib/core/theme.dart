import 'package:flutter/material.dart';

/// Design tokens — theme "Sunset Journey" (owner decision 2026-10-25), brand "Sabai jai Travel".
/// Primary #C2410C · Secondary #FB923C · Background #FFF8F3 · Surface #FFFFFF · Accent #EAB308 · Text #431D12.
class TepColors {
  static const primary = Color(0xFFC2410C); // headers, primary buttons, active icons (white text: 5.2:1)
  static const primaryDeep = Color(0xFF9A3412); // gradient start / pressed
  static const secondary = Color(0xFFFB923C); // highlights; use ink text on it, never white
  static const background = Color(0xFFFFF8F3);
  static const card = Colors.white;
  static const accent = Color(0xFFEAB308); // rating stars
  static const ink = Color(0xFF431D12); // body text
  static const muted = Color(0xFF7C5A4B); // secondary text (>= 4.5:1 on background)
  static const line = Color(0xFFF1DFD3);
  static const success = Color(0xFF23925D);
  static const danger = Color(0xFFB91C1C);
  static const partner = Color(0xFF3988A2); // portal preview only
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
      textTheme: ThemeData.light().textTheme.apply(bodyColor: TepColors.ink, displayColor: TepColors.ink),
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
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
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
