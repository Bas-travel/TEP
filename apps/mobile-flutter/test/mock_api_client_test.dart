import 'package:flutter_test/flutter_test.dart';
import 'package:tep_mobile/data/api_client.dart';
import 'package:tep_mobile/data/mock_api_client.dart';
import 'package:tep_mobile/data/models.dart';

MockApiClient fresh() => MockApiClient(latency: Duration.zero, settleAfter: Duration.zero);

void main() {
  test('rejects unauthenticated calls and bad credentials', () async {
    final api = fresh();
    await expectLater(api.wallet(), throwsA(isA<ApiException>().having((e) => e.status, 'status', 401)));
    await expectLater(api.login('demo@tep.dev', 'wrong'), throwsA(isA<ApiException>()));
    await api.login(MockApiClient.demoEmail, MockApiClient.demoPassword);
    expect((await api.me()).firstName, 'สวัสดี');
  });

  test('catalog search filters by type, query, destination and paginates', () async {
    final api = fresh();
    final hotels = await api.listings(type: 'hotel');
    expect(hotels.items.every((l) => l.type == 'hotel'), isTrue);
    expect(hotels.items.length, 5);
    final phuket = await api.listings(destinationId: 1);
    expect(phuket.items.length, 4);
    expect((await api.listings(q: 'kata')).items.single.id, 2);
    final p1 = await api.listings(limit: 5);
    expect(p1.items.length, 5);
    expect(p1.nextCursor, '5');
    final p2 = await api.listings(limit: 5, cursor: p1.nextCursor);
    expect(p2.items.length, 5);
    expect({...p1.items.map((e) => e.id), ...p2.items.map((e) => e.id)}.length, 10);
  });

  test('booking priced like the API: subtotal + 3% fee, WELCOME10 discount', () async {
    final api = fresh();
    await api.login(MockApiClient.demoEmail, MockApiClient.demoPassword);
    final b = await api.createBooking(lines: [const BookingLine(1, 1)], idempotencyKey: 'a');
    expect(b.subtotal, 3200);
    expect(b.serviceFee, 96);
    expect(b.amount, 3296);
    expect(b.status, 'pending_payment');
    final b2 = await api.createBooking(lines: [const BookingLine(1, 1)], couponCode: 'welcome10', idempotencyKey: 'b');
    expect(b2.amount, 3200 + 96 - 320);
    await expectLater(api.createBooking(lines: [const BookingLine(6, 1)], couponCode: 'WELCOME10', idempotencyKey: 'c'), throwsA(isA<ApiException>()));
  });

  test('idempotency key replays the same booking and intent', () async {
    final api = fresh();
    await api.login(MockApiClient.demoEmail, MockApiClient.demoPassword);
    final a = await api.createBooking(lines: [const BookingLine(7, 1)], idempotencyKey: 'same');
    final b = await api.createBooking(lines: [const BookingLine(7, 1)], idempotencyKey: 'same');
    expect(b.id, a.id);
    expect((await api.bookings()).length, 1);
    final i1 = await api.createIntent(bookingId: a.id, method: 'wallet', idempotencyKey: 'pay');
    final i2 = await api.createIntent(bookingId: a.id, method: 'wallet', idempotencyKey: 'pay');
    expect(i2.id, i1.id);
    expect((await api.wallet()).balance, closeTo(1500 - a.amount, 0.001)); // charged once
  });

  test('wallet payment confirms booking and debits; cancel refunds', () async {
    final api = fresh();
    await api.login(MockApiClient.demoEmail, MockApiClient.demoPassword);
    final b = await api.createBooking(lines: [const BookingLine(7, 1)], idempotencyKey: 'k1'); // 990 + 29.70
    final pi = await api.createIntent(bookingId: b.id, method: 'wallet', idempotencyKey: 'p1');
    expect(pi.status, 'succeeded');
    expect((await api.booking(b.id)).status, 'confirmed');
    expect((await api.wallet()).balance, closeTo(1500 - b.amount, 0.001));
    final c = await api.cancelBooking(b.id, 'เปลี่ยนแผน');
    expect(c.status, 'cancelled');
    expect(c.refundAmount, b.amount);
    expect((await api.wallet()).balance, closeTo(1500, 0.001));
  });

  test('wallet short returns 402 WALLET_INSUFFICIENT', () async {
    final api = fresh();
    await api.login(MockApiClient.demoEmail, MockApiClient.demoPassword);
    final b = await api.createBooking(lines: [const BookingLine(8, 1)], idempotencyKey: 'k2'); // 4200+
    await expectLater(
      api.createIntent(bookingId: b.id, method: 'wallet', idempotencyKey: 'p2'),
      throwsA(isA<ApiException>().having((e) => e.status, 'status', 402).having((e) => e.isWalletShort, 'short', isTrue)),
    );
  });

  test('promptpay settles to confirmed; top-up credits wallet after settle', () async {
    final api = fresh();
    await api.login(MockApiClient.demoEmail, MockApiClient.demoPassword);
    final b = await api.createBooking(lines: [const BookingLine(1, 1)], idempotencyKey: 'k3');
    final pi = await api.createIntent(bookingId: b.id, method: 'promptpay', idempotencyKey: 'p3');
    expect(pi.status, 'pending');
    expect(pi.promptPayQrUrl, isNotNull);
    expect((await api.intent(pi.id)).succeeded, isTrue); // settleAfter = 0
    expect((await api.booking(b.id)).status, 'confirmed');
    await api.topUp(500);
    expect((await api.wallet()).balance, 2000);
  });
}
