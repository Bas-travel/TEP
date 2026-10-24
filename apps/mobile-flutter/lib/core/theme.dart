import 'package:flutter/material.dart';

/// Design tokens sampled from the "Travel Super App" board (docs/ux/board-mapping.md).
class TepColors {
  static const navy = Color(0xFF002765); // header
  static const navyBar = Color(0xFF082D66); // nav bar / primary buttons
  static const background = Color(0xFFF4F8FC);
  static const success = Color(0xFF23925D);
  static const partner = Color(0xFF3988A2);
  static const accent = Color(0xFFF5A524); // rating star
  static const danger = Color(0xFFD64545);
  static const ink = Color(0xFF0F1B33);
  static const muted = Color(0xFF64748B);
  static const line = Color(0xFFE2E8F0);
  static const card = Colors.white;
}

class TepTheme {
  static const _fallback = ['Noto Sans Thai', 'Sarabun', 'Roboto'];

  static ThemeData light() {
    final scheme = ColorScheme.fromSeed(
      seedColor: TepColors.navyBar,
      primary: TepColors.navyBar,
      secondary: TepColors.success,
      surface: TepColors.card,
      error: TepColors.danger,
    );
    return ThemeData(
      useMaterial3: true,
      colorScheme: scheme,
      scaffoldBackgroundColor: TepColors.background,
      fontFamilyFallback: _fallback,
      appBarTheme: const AppBarTheme(
        backgroundColor: TepColors.navy,
        foregroundColor: Colors.white,
        elevation: 0,
        centerTitle: true,
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: TepColors.navyBar,
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
        indicatorColor: TepColors.navyBar.withValues(alpha: 0.10),
      ),
    );
  }
}
