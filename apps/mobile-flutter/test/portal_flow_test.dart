import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tep_mobile/data/mock_api_client.dart';
import 'package:tep_mobile/main.dart';

Future<void> _bootAs(WidgetTester tester, String email) async {
  tester.view.physicalSize = const Size(1080, 2400);
  tester.view.devicePixelRatio = 2.5;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(TepApp(api: MockApiClient(latency: Duration.zero, settleAfter: Duration.zero)));
  await tester.pumpAndSettle();
  await tester.enterText(find.byKey(const Key('email')), email);
  await tester.tap(find.byKey(const Key('submit')));
  await tester.pumpAndSettle();
}

Future<void> _tab(WidgetTester tester, String label) async {
  await tester.tap(find.descendant(of: find.byKey(const Key('portal-nav')), matching: find.text(label)));
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('staff sees the staff portal; every menu shows its own page', (tester) async {
    await _bootAs(tester, 'staff@tep.dev');
    expect(find.text('พนักงาน · ภาพรวม'), findsOneWidget);
    expect(find.text('การจองรอตรวจ'), findsOneWidget);
    await _tab(tester, 'ตรวจการจอง');
    expect(find.text('พนักงาน · ตรวจการจอง'), findsOneWidget);
    expect(find.byKey(const Key('filter-รอตรวจ')), findsOneWidget);
    expect(find.text('Andaman Sunset Resort'), findsOneWidget);
    await _tab(tester, 'ช่วยเหลือ');
    expect(find.text('สอบถามการยกเลิกการจอง'), findsOneWidget);
    await _tab(tester, 'รายงาน');
    expect(find.textContaining('รายงานประจำวัน'), findsOneWidget);
  });

  testWidgets('staff replies to a ticket from the ticket page', (tester) async {
    await _bootAs(tester, 'staff@tep.dev');
    await _tab(tester, 'ช่วยเหลือ');
    await tester.tap(find.text('สอบถามการยกเลิกการจอง'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('reply-text')), 'ยกเลิกได้ที่เมนูการจองค่ะ');
    await tester.tap(find.byKey(const Key('reply-send')));
    await tester.pumpAndSettle();
    expect(find.text('ยกเลิกได้ที่เมนูการจองค่ะ'), findsOneWidget);
    expect(find.text('ตอบแล้ว'), findsWidgets);
  });

  testWidgets('partner portal: dashboard, listings, bookings, earnings', (tester) async {
    await _bootAs(tester, 'partner@tep.dev');
    expect(find.text('พาร์ทเนอร์ · ภาพรวม'), findsOneWidget);
    expect(find.text('Andaman Partner Co.'), findsOneWidget);
    await _tab(tester, 'สินค้า/บริการ');
    expect(find.byKey(const Key('listing-add')), findsOneWidget);
    expect(find.text('Andaman Sunset Resort'), findsOneWidget);
    await _tab(tester, 'การจอง');
    expect(find.byKey(const Key('complete-1')), findsOneWidget);
    await tester.tap(find.byKey(const Key('complete-1')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('dialog-confirm')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('complete-1')), findsNothing);
    await _tab(tester, 'รายได้');
    expect(find.byKey(const Key('profile-edit')), findsOneWidget);
  });

  testWidgets('admin portal: users, partner approval, system', (tester) async {
    await _bootAs(tester, 'admin@tep.dev');
    expect(find.text('แอดมิน · ภาพรวม'), findsOneWidget);
    await _tab(tester, 'ผู้ใช้');
    expect(find.text('สวัสดี อนันต์'), findsOneWidget);
    await _tab(tester, 'พาร์ทเนอร์');
    expect(find.text('Fah Local Tours'), findsOneWidget);
    await tester.tap(find.byKey(const Key('approve-2')));
    await tester.pumpAndSettle();
    expect(find.text('อนุมัติแล้ว'), findsWidgets);
    await _tab(tester, 'ระบบ');
    expect(find.text('สถานะระบบ'), findsOneWidget);
    await tester.tap(find.byKey(const Key('tab-audit')));
    await tester.pumpAndSettle();
    expect(find.text('admin.partner.update'), findsOneWidget);
  });

  testWidgets('traveler keeps the traveler app and cannot reach a portal', (tester) async {
    await _bootAs(tester, 'demo@tep.dev');
    expect(find.byType(NavigationBar), findsOneWidget);
    expect(find.byKey(const Key('portal-nav')), findsNothing);
  });
}
