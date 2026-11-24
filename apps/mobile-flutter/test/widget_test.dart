// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tep_mobile/data/mock_api_client.dart';
import 'package:tep_mobile/main.dart';

void main() {
  testWidgets('desktop login screen exposes TEP desktop branding',
      (tester) async {
    tester.view.physicalSize = const Size(1440, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.reset);

    await tester.pumpWidget(
        TepApp(api: MockApiClient(latency: Duration.zero), modeLabel: 'mock'));
    await tester.pumpAndSettle();

    expect(find.text('Your World, One App.'), findsOneWidget);
    expect(find.text('Discover places. Plan adventures. Enjoy every journey.'),
        findsOneWidget);
  });

  testWidgets('desktop shell uses navigation rail after login', (tester) async {
    tester.view.physicalSize = const Size(1440, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.reset);

    await tester.pumpWidget(
        TepApp(api: MockApiClient(latency: Duration.zero), modeLabel: 'mock'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('submit')));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('desktop-navigation')), findsOneWidget);
    expect(find.text('หน้าแรก'), findsWidgets);
    expect(find.byType(NavigationBar), findsNothing);
  });

  for (final demo in const [
    ('demo-admin', 'admin@tep.dev', 'แอดมิน'),
    ('demo-partner', 'partner@tep.dev', 'พาร์ทเนอร์'),
    ('demo-staff', 'staff@tep.dev', 'พนักงาน'),
  ]) {
    testWidgets('${demo.$1} quick login routes to its own portal',
        (tester) async {
      tester.view.physicalSize = const Size(1440, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.reset);
      await tester.pumpWidget(TepApp(
          api: MockApiClient(latency: Duration.zero), modeLabel: 'mock'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byKey(const Key('email')), demo.$2);
      await tester.enterText(find.byKey(const Key('password')), 'Passw0rd!');
      await tester.tap(find.byKey(const Key('submit')));
      await tester.pumpAndSettle();
      final title = tester.widget<Text>(find.byKey(const Key('portal-title')));
      expect(title.data, startsWith(demo.$3));
      expect(find.byKey(const Key('portal-nav')), findsOneWidget);
    });
  }

  testWidgets('moderator has no portal and uses the traveler app',
      (tester) async {
    tester.view.physicalSize = const Size(1440, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(
        TepApp(api: MockApiClient(latency: Duration.zero), modeLabel: 'mock'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('email')), 'moderator@tep.dev');
    await tester.enterText(find.byKey(const Key('password')), 'Passw0rd!');
    await tester.tap(find.byKey(const Key('submit')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('portal-nav')), findsNothing);
    expect(find.byKey(const Key('desktop-navigation')), findsOneWidget);
  });
}
