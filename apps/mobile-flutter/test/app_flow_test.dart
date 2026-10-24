import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tep_mobile/data/mock_api_client.dart';
import 'package:tep_mobile/main.dart';

Future<void> _boot(WidgetTester tester) async {
  tester.view.physicalSize = const Size(1080, 2400);
  tester.view.devicePixelRatio = 2.5;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(TepApp(api: MockApiClient(latency: Duration.zero, settleAfter: Duration.zero)));
  await tester.pumpAndSettle();
}

Future<void> _tab(WidgetTester tester, String label) async {
  await tester.tap(find.descendant(of: find.byType(NavigationBar), matching: find.text(label)));
  await tester.pumpAndSettle();
}

Future<void> _login(WidgetTester tester) async {
  await tester.tap(find.byKey(const Key('submit')));
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('wrong password shows error and stays on login', (tester) async {
    await _boot(tester);
    await tester.enterText(find.byKey(const Key('password')), 'bad-password');
    await _login(tester);
    expect(find.text('เข้าสู่ระบบ'), findsWidgets);
    expect(find.byType(NavigationBar), findsNothing);
  });

  testWidgets('login lands on Home with greeting, destinations and listings', (tester) async {
    await _boot(tester);
    await _login(tester);
    expect(find.byType(NavigationBar), findsOneWidget);
    expect(find.text('สวัสดี, สวัสดี'), findsOneWidget);
    expect(find.text('จุดหมายยอดนิยม'), findsOneWidget);
    expect(find.text('แนะนำสำหรับคุณ'), findsOneWidget);
  });

  testWidgets('category chip opens filtered search results', (tester) async {
    await _boot(tester);
    await _login(tester);
    await tester.tap(find.byKey(const Key('cat-hotel')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('results')), findsOneWidget);
    expect(find.text('Andaman Sunset Resort'), findsOneWidget);
    expect(find.text('ครัวทะเลเก่าภูเก็ต'), findsNothing); // restaurant filtered out
  });

  testWidgets('full journey: category → detail → book → pay with Wallet → success → listed in My Bookings', (tester) async {
    await _boot(tester);
    await _login(tester);
    await tester.tap(find.byKey(const Key('cat-attraction')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('ดอยสุเทพ + ไนท์ซาฟารี'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('book-now')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('to-payment')));
    await tester.pumpAndSettle();
    expect(find.textContaining('TEP-'), findsWidgets);
    await tester.tap(find.byKey(const Key('method-wallet')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('pay')));
    await tester.pumpAndSettle();
    expect(find.text('การจองสำเร็จ'), findsOneWidget);
    await tester.tap(find.byKey(const Key('done')));
    await tester.pumpAndSettle();
    await _tab(tester, 'การจอง');
    expect(find.text('ดอยสุเทพ + ไนท์ซาฟารี'), findsOneWidget);
    expect(find.text('ยืนยันแล้ว'), findsWidgets);
  });

  testWidgets('wallet tab shows seeded balance', (tester) async {
    await _boot(tester);
    await _login(tester);
    await _tab(tester, 'Wallet');
    expect(find.text('฿1,500'), findsOneWidget);
  });

  testWidgets('logout returns to login', (tester) async {
    await _boot(tester);
    await _login(tester);
    await _tab(tester, 'โปรไฟล์');
    expect(find.byKey(const Key('profile-name')), findsOneWidget);
    await tester.tap(find.byKey(const Key('logout')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('submit')), findsOneWidget);
  });
}
