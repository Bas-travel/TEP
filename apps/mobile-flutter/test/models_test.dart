import 'package:flutter_test/flutter_test.dart';
import 'package:tep_mobile/core/format.dart';
import 'package:tep_mobile/data/models.dart';

void main() {
  test('thb formats whole and fractional amounts with grouping', () {
    expect(thb(0), '฿0');
    expect(thb(3200), '฿3,200');
    expect(thb(1234567), '฿1,234,567');
    expect(thb(99.5), '฿99.50');
    expect(thb(3296.0), '฿3,296');
  });

  test('gradientFromCss reads hex colors, falls back when absent', () {
    final g = gradientFromCss('linear-gradient(135deg,#0b5ed7,#2bb3c0)');
    expect(g.colors.length, 2);
    expect(g.colors.first.toARGB32(), 0xFF0B5ED7);
    expect(gradientFromCss('nope').colors.length, 2);
  });

  test('ListingDetail.fromJson maps API fields', () {
    final d = ListingDetail.fromJson({
      'id': 1,
      'type': 'hotel',
      'name': 'A',
      'location': 'B',
      'rating': 4,
      'reviewCount': 10,
      'priceFrom': 3200,
      'description': 'x',
      'amenities': ['Wi-Fi'],
      'mapQuery': 'q',
    });
    expect(d.rating, 4.0);
    expect(d.priceFrom, 3200.0);
    expect(d.amenities, ['Wi-Fi']);
    expect(d.imageUrl, isNull);
  });

  test('ApiException.fromJson reads RFC7807 problem with field errors', () {
    final e = ApiException.fromJson(422, {
      'title': 'Validation failed',
      'code': 'VALIDATION_FAILED',
      'requestId': 'r1',
      'errors': [
        {'field': 'couponCode', 'message': 'bad coupon'}
      ],
    });
    expect(e.status, 422);
    expect(e.code, 'VALIDATION_FAILED');
    expect(e.errors.single.field, 'couponCode');
    expect(e.userMessage, 'bad coupon');
    expect(ApiException.fromJson(402, {'code': 'WALLET_INSUFFICIENT', 'title': 'x'}).isWalletShort, isTrue);
  });

  test('BookingDetail.fromJson keeps nullable refund', () {
    final b = BookingDetail.fromJson({
      'id': 1,
      'bookingNumber': 'TEP-1',
      'type': 'hotel',
      'title': 't',
      'subtitle': 's',
      'amount': 3296,
      'method': '',
      'date': '2026-10-25T01:02:03Z',
      'status': 'pending_payment',
      'subtotal': 3200,
      'serviceFee': 96,
    });
    expect(b.refundAmount, isNull);
    expect(b.date.isUtc, isTrue);
  });
}
