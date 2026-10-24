import 'package:flutter/material.dart';

/// 12 500 -> "฿12,500", 99.5 -> "฿99.50".
String thb(num v) {
  final whole = v == v.roundToDouble();
  final parts = v.toStringAsFixed(whole ? 0 : 2).split('.');
  final grouped = parts[0].replaceAllMapped(RegExp(r'\B(?=(\d{3})+(?!\d))'), (_) => ',');
  return '฿${parts.length > 1 ? '$grouped.${parts[1]}' : grouped}';
}

String dmy(DateTime d) {
  final l = d.toLocal();
  return '${l.day.toString().padLeft(2, '0')}/${l.month.toString().padLeft(2, '0')}/${l.year}';
}

/// Parses `linear-gradient(135deg,#0b5ed7,#2bb3c0)` (as served in Destination.gradientCss).
LinearGradient gradientFromCss(String css) {
  final colors = RegExp(r'#([0-9a-fA-F]{6})')
      .allMatches(css)
      .map((m) => Color(int.parse('FF${m.group(1)}', radix: 16)))
      .toList();
  if (colors.length < 2) {
    return const LinearGradient(colors: [Color(0xFF002765), Color(0xFF3988A2)]);
  }
  return LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: colors);
}

const typeLabels = {
  'hotel': 'โรงแรม',
  'restaurant': 'ร้านอาหาร',
  'attraction': 'ที่เที่ยว',
  'flight': 'เที่ยวบิน',
  'trip': 'แพ็กเกจ',
};

IconData typeIcon(String type) => switch (type) {
      'hotel' => Icons.hotel_rounded,
      'restaurant' => Icons.restaurant_rounded,
      'attraction' => Icons.attractions_rounded,
      'flight' => Icons.flight_rounded,
      'trip' => Icons.luggage_rounded,
      _ => Icons.place_rounded,
    };

const statusLabels = {
  'pending_payment': 'รอชำระเงิน',
  'confirmed': 'ยืนยันแล้ว',
  'completed': 'เสร็จสิ้น',
  'cancelled': 'ยกเลิกแล้ว',
};
