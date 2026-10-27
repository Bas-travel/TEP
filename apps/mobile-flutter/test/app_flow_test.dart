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
  testWidgets('main menu lists all 12 board entries and opens Coupons', (tester) async {
    await _boot(tester);
    await _login(tester);
    await tester.tap(find.byKey(const Key('home-menu')));
    await tester.pumpAndSettle();
    for (final id in ['home', 'search', 'bookings', 'trips', 'tickets', 'wallet', 'coupons', 'notifications', 'messages', 'profile', 'settings', 'support']) {
      expect(find.byKey(Key('menu-$id')), findsOneWidget, reason: 'menu entry $id');
    }
    await tester.tap(find.byKey(const Key('menu-coupons')));
    await tester.pumpAndSettle();
    expect(find.text('คูปองของฉัน'), findsOneWidget);
    expect(find.text('ส่วนลดสมาชิกใหม่ 10%'), findsOneWidget);
  });

  testWidgets('menu entry for Wallet switches the bottom tab', (tester) async {
    await _boot(tester);
    await _login(tester);
    await tester.tap(find.byKey(const Key('home-menu')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('menu-wallet')));
    await tester.pumpAndSettle();
    expect(find.text('฿1,500'), findsOneWidget);
  });

  Future<void> openCommunity(WidgetTester tester) async {
    await tester.tap(find.byKey(const Key('home-menu')));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(find.byKey(const Key('menu-community')), 200);
    await tester.tap(find.byKey(const Key('menu-community')));
    await tester.pumpAndSettle();
  }

  testWidgets('community: write a post, open it, comment, and save it', (tester) async {
    await _boot(tester);
    await _login(tester);
    await openCommunity(tester);
    await tester.tap(find.byKey(const Key('new-post')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('post-text')), 'สั้น');
    await tester.tap(find.byKey(const Key('post-submit')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('post-text')), findsOneWidget); // 422: sheet stays open
    await tester.enterText(find.byKey(const Key('post-dest')), 'ภูเก็ต');
    await tester.enterText(find.byKey(const Key('post-text')), 'ทะเลภูเก็ตช่วงเช้าสงบมาก แนะนำมากๆ');
    await tester.tap(find.byKey(const Key('post-submit')));
    await tester.pumpAndSettle();
    expect(find.text('ทะเลภูเก็ตช่วงเช้าสงบมาก แนะนำมากๆ'), findsOneWidget);
    await tester.tap(find.byKey(const Key('save-6')));
    await tester.pumpAndSettle();
    expect(find.byIcon(Icons.bookmark_rounded), findsOneWidget);
    await tester.tap(find.byKey(const Key('post-6')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('comment-text')), 'ขอบคุณที่แชร์ครับ');
    await tester.tap(find.byKey(const Key('comment-send')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('comment-1')), findsOneWidget);
    expect(find.text('ขอบคุณที่แชร์ครับ'), findsOneWidget);
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(tester.widget<Text>(find.byKey(const Key('comment-count-6'))).data, '1');
  });

  testWidgets('community groups: join and leave updates the member count', (tester) async {
    await _boot(tester);
    await _login(tester);
    await openCommunity(tester);
    await tester.tap(find.byKey(const Key('open-groups')));
    await tester.pumpAndSettle();
    expect(tester.widget<Text>(find.byKey(const Key('group-members-2'))).data, '1 สมาชิก');
    await tester.tap(find.byKey(const Key('group-join-2')));
    await tester.pumpAndSettle();
    expect(tester.widget<Text>(find.byKey(const Key('group-members-2'))).data, '2 สมาชิก');
    expect(find.text('ออกจากกลุ่ม'), findsOneWidget);
    await tester.tap(find.byKey(const Key('group-join-2')));
    await tester.pumpAndSettle();
    expect(tester.widget<Text>(find.byKey(const Key('group-members-2'))).data, '1 สมาชิก');
  });

  testWidgets('change password: wrong current stays, correct signs out; new password works', (tester) async {
    await _boot(tester);
    await _login(tester);
    await _tab(tester, 'โปรไฟล์');
    await tester.tap(find.byKey(const Key('open-change-password')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('pw-current')), 'wrong-password');
    await tester.enterText(find.byKey(const Key('pw-new')), 'NewPassw0rd!');
    await tester.tap(find.byKey(const Key('pw-submit')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('pw-submit')), findsOneWidget);
    await tester.enterText(find.byKey(const Key('pw-current')), 'Passw0rd!');
    await tester.tap(find.byKey(const Key('pw-submit')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('submit')), findsOneWidget); // back on the login screen
    await tester.enterText(find.byKey(const Key('password')), 'Passw0rd!');
    await _login(tester);
    expect(find.byType(NavigationBar), findsNothing); // old password rejected
    await tester.enterText(find.byKey(const Key('password')), 'NewPassw0rd!');
    await _login(tester);
    expect(find.byType(NavigationBar), findsOneWidget);
  });

  testWidgets('delete account asks for confirmation, then signs out', (tester) async {
    await _boot(tester);
    await _login(tester);
    await _tab(tester, 'โปรไฟล์');
    await tester.tap(find.byKey(const Key('delete-account')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('delete-cancel')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('profile-name')), findsOneWidget);
    await tester.tap(find.byKey(const Key('delete-account')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('delete-confirm')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('submit')), findsOneWidget);
  });
}
