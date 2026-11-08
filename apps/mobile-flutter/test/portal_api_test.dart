import 'package:flutter_test/flutter_test.dart';
import 'package:tep_mobile/data/api_client.dart';
import 'package:tep_mobile/data/mock_api_client.dart';

Future<MockApiClient> as_(String email) async {
  final api = MockApiClient(latency: Duration.zero, settleAfter: Duration.zero);
  await api.login(email, MockApiClient.demoPassword);
  return api;
}

Matcher status(int s) => throwsA(isA<ApiException>().having((e) => e.status, 'status', s));
List<Json> items(Json j) => (j['items'] as List).cast<Json>();

void main() {
  test('each portal rejects the other roles with 403', () async {
    final traveler = await as_('demo@tep.dev');
    final partner = await as_('partner@tep.dev');
    final staff = await as_('staff@tep.dev');
    final admin = await as_('admin@tep.dev');
    for (final path in ['/partner/v1/dashboard', '/admin/v1/dashboard', '/staff/v1/dashboard']) {
      await expectLater(traveler.portal('GET', path), status(403));
    }
    await expectLater(staff.portal('GET', '/admin/v1/users'), status(403));
    await expectLater(partner.portal('GET', '/staff/v1/tickets'), status(403));
    await expectLater(admin.portal('GET', '/partner/v1/listings'), status(403));
    expect((await admin.portal('GET', '/staff/v1/dashboard'))['openTickets'], 1); // admin may use the staff portal
  });

  test('partner manages own listings and bookings', () async {
    final api = await as_('partner@tep.dev');
    expect(items(await api.portal('GET', '/partner/v1/listings')).length, 4);
    await expectLater(api.portal('GET', '/partner/v1/listings/5'), status(404)); // someone else's
    await expectLater(api.portal('POST', '/partner/v1/listings', body: {'name': '', 'type': 'flight', 'priceFrom': -1, 'destinationId': 1}), status(422));
    final created = await api.portal('POST', '/partner/v1/listings', body: {
      'name': 'Sunset Villa', 'type': 'hotel', 'location': 'Phuket', 'destinationId': 1, 'priceFrom': 2500, 'status': 'published',
    });
    expect(created['status'], 'published');
    final patched = await api.portal('PATCH', '/partner/v1/listings/${created['id']}', body: {'status': 'archived'});
    expect(patched['status'], 'archived');

    final all = await api.portal('GET', '/partner/v1/bookings');
    expect(all['total'], 3);
    final done = await api.portal('POST', '/partner/v1/bookings/1/complete');
    expect(done['status'], 'completed');
    await expectLater(api.portal('POST', '/partner/v1/bookings/1/complete'), status(409));
    await expectLater(api.portal('POST', '/partner/v1/bookings/3/reject', body: {}), status(422));
    final rejected = await api.portal('POST', '/partner/v1/bookings/3/reject', body: {'reason': 'ห้องเต็ม'});
    expect(rejected['status'], 'cancelled');

    final e = await api.portal('GET', '/partner/v1/earnings');
    expect(e['commissionRate'], 12);
    expect((e['commission'] as num).toDouble(), closeTo((e['gross'] as num) * 0.12, 0.011));
  });

  test('admin approves an application, which promotes the owner; transitions are enforced', () async {
    final api = await as_('admin@tep.dev');
    await expectLater(api.portal('PATCH', '/admin/v1/partners/2', body: {'status': 'suspended'}), status(409));
    await expectLater(api.portal('PATCH', '/admin/v1/partners/2', body: {'status': 'rejected'}), status(422));
    await expectLater(api.portal('PATCH', '/admin/v1/partners/2', body: {'commissionRate': 60}), status(422));
    final p = await api.portal('PATCH', '/admin/v1/partners/2', body: {'status': 'approved', 'commissionRate': 10});
    expect(p['status'], 'approved');
    expect((await api.portal('GET', '/admin/v1/users/4'))['role'], 'partner');
    // users: never to partner, never self
    await expectLater(api.portal('PATCH', '/admin/v1/users/1', body: {'role': 'partner'}), status(422));
    await expectLater(api.portal('PATCH', '/admin/v1/users/6', body: {'status': 'suspended'}), status(409));
    final u = await api.portal('PATCH', '/admin/v1/users/1', body: {'role': 'moderator'});
    expect(u['role'], 'moderator');
    final audit = await api.portal('GET', '/admin/v1/audit-logs', query: {'action': 'admin.'});
    expect(audit['total'], 2);
  });

  test('staff verifies bookings and answers tickets; traveler sees the answer', () async {
    final staff = await as_('staff@tep.dev');
    await expectLater(staff.portal('POST', '/staff/v1/bookings/1/check', body: {'result': 'flagged'}), status(422));
    await expectLater(staff.portal('POST', '/staff/v1/bookings/3/check', body: {'result': 'verified'}), status(409));
    final b = await staff.portal('POST', '/staff/v1/bookings/1/check', body: {'result': 'verified'});
    expect(b['verification'], 'verified');
    expect((await staff.portal('GET', '/staff/v1/dashboard'))['uncheckedBookings'], 0);

    final t = await staff.portal('POST', '/staff/v1/tickets/1/replies', body: {'message': 'รับเรื่องแล้วค่ะ'});
    expect(t['status'], 'answered');
    final r = await staff.portal('PATCH', '/staff/v1/tickets/1', body: {'status': 'resolved'});
    expect(r['status'], 'resolved');
    await expectLater(staff.portal('POST', '/staff/v1/tickets/1/replies', body: {'message': 'x'}), status(409));
    expect((await staff.portal('GET', '/staff/v1/reports/daily'))['bookingsVerified'], 1);
  });

  test('traveler applies, ticket replies reopen the ticket', () async {
    final api = await as_('demo@tep.dev');
    await expectLater(api.portal('GET', '/v1/partners/me'), status(404));
    expect((await api.portal('POST', '/v1/partners/apply', body: {'businessName': 'Demo Tours'}))['status'], 'pending');
    await expectLater(api.portal('POST', '/v1/partners/apply', body: {'businessName': 'Again'}), status(409));
    final t = await api.createTicket(subject: 'ช่วยด้วย', message: 'ขอเลื่อนวัน');
    expect(t.status, 'open');
    final out = await api.portal('POST', '/v1/support/tickets/${t.id}/replies', body: {'message': 'วันที่ 20'});
    expect(out['status'], 'open');
    expect((await api.tickets()).first.replies.length, 1);
  });
}
