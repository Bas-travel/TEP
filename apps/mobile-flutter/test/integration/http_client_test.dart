@TestOn('vm')
library;

import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:tep_mobile/data/api_client.dart';
import 'package:tep_mobile/data/http_api_client.dart';
import 'package:tep_mobile/data/models.dart';

/// Contract test: the real HttpApiClient against a running services/api-go (APP_ENV=dev).
/// Run: TEP_API_URL=http://127.0.0.1:8080 flutter test test/integration
final _url = Platform.environment['TEP_API_URL'];

void main() {
  group('HttpApiClient ↔ services/api-go', () {
    late HttpApiClient api;
    setUp(() => api = HttpApiClient(baseUrl: _url!));

    test('catalog is public and shaped as the mock fixtures', () async {
      final d = await api.destinations();
      expect(d.length, 6);
      final page = await api.listings(type: 'hotel');
      expect(page.items.length, 5);
      expect((await api.listing(1)).amenities, isNotEmpty);
      await expectLater(api.listing(9999), throwsA(isA<ApiException>().having((e) => e.status, 'status', 404)));
      expect((await api.faqs()).length, 4);
    });

    test('auth errors map to ApiException with code', () async {
      await expectLater(api.wallet(), throwsA(isA<ApiException>().having((e) => e.status, 'status', 401)));
      await expectLater(api.login('demo@tep.dev', 'nope'), throwsA(isA<ApiException>().having((e) => e.code, 'code', 'AUTH_UNAUTHENTICATED')));
    });

    test('booking round-trips startDate / endDate / guests and rejects a past date', () async {
      await api.login('demo@tep.dev', 'Passw0rd!');
      final start = DateTime.now().add(const Duration(days: 5));
      final b = await api.createBooking(
          lines: [const BookingLine(1, 1)], startDate: start, endDate: start.add(const Duration(days: 2)), guests: 2, idempotencyKey: 'it-dates-${DateTime.now().microsecondsSinceEpoch}');
      final read = await api.booking(b.id);
      expect(read.guests, 2);
      expect(read.startDate, isNotNull);
      expect(read.endDate, isNotNull);
      await expectLater(
          api.createBooking(lines: [const BookingLine(1, 1)], startDate: DateTime.now().subtract(const Duration(days: 4)), idempotencyKey: 'it-past-${DateTime.now().microsecondsSinceEpoch}'),
          throwsA(isA<ApiException>().having((e) => e.status, 'status', 422)));
    });

    test('book → pay with wallet → cancel → refund (full journey)', () async {
      await api.login('demo@tep.dev', 'Passw0rd!');
      expect((await api.me()).email, 'demo@tep.dev');
      final before = (await api.wallet()).balance;
      final b = await api.createBooking(lines: [const BookingLine(7, 1)], idempotencyKey: 'it-${DateTime.now().microsecondsSinceEpoch}');
      expect(b.status, 'pending_payment');
      final pi = await api.createIntent(bookingId: b.id, method: 'wallet', idempotencyKey: 'pi-${DateTime.now().microsecondsSinceEpoch}');
      expect(pi.succeeded, isTrue);
      expect((await api.booking(b.id)).status, 'confirmed');
      expect((await api.wallet()).balance, closeTo(before - b.amount, 0.001));
      final c = await api.cancelBooking(b.id, 'integration test');
      expect(c.status, 'cancelled');
      expect((await api.wallet()).balance, closeTo(before, 0.001));
      await api.logout();
    });
  }, skip: _url == null ? 'set TEP_API_URL to run against a live API' : false);
}
