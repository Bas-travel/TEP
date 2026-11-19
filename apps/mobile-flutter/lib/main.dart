import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import 'app_scope.dart';
import 'core/theme.dart';
import 'data/api_client.dart';
import 'data/http_api_client.dart';
import 'data/mock_api_client.dart';
import 'data/models.dart';
import 'ui/login_screen.dart';
import 'ui/portal/portal_shell.dart';
import 'ui/shell.dart';

/// `flutter run --dart-define=API_MODE=http --dart-define=API_BASE_URL=http://10.0.2.2:8080`
const _apiMode = String.fromEnvironment('API_MODE', defaultValue: 'mock');
const _apiBaseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: '');

void main() {
  final ApiClient api;
  if (_apiMode == 'http') {
    final baseUrl = _apiBaseUrl.isNotEmpty
        ? _apiBaseUrl
        : kIsWeb
            ? 'http://localhost:8080'
            : 'http://10.0.2.2:8080';
    api = HttpApiClient(baseUrl: baseUrl);
  } else {
    api = MockApiClient();
  }
  runApp(TepApp(api: api, modeLabel: _apiMode));
}

class TepApp extends StatefulWidget {
  const TepApp({super.key, required this.api, this.modeLabel = 'mock'});
  final ApiClient api;
  final String modeLabel;

  @override
  State<TepApp> createState() => _TepAppState();
}

class _TepAppState extends State<TepApp> {
  late final SessionController _session = SessionController(widget.api);

  @override
  void dispose() {
    _session.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AppScope(
      session: _session,
      child: MaterialApp(
        title: 'TEP — Travel Super App',
        debugShowCheckedModeBanner: false,
        theme: TepTheme.light(),
        home: ListenableBuilder(
          listenable: _session,
          builder: (context, _) => _session.signedIn
              ? _authenticatedShell(_session, widget.modeLabel)
              : LoginScreen(modeLabel: widget.modeLabel),
        ),
        scrollBehavior: const MaterialScrollBehavior().copyWith(
          scrollbars: true,
        ),
      ),
    );
  }

  /// Back-office roles get their own functional portal; travelers and
  /// moderators (no portal in the MVP scope) use the traveler app.
  Widget _authenticatedShell(SessionController session, String modeLabel) {
    final role = session.user?.role ?? AccessRole.traveler;
    if (const {AccessRole.partner, AccessRole.admin, AccessRole.staff}
        .contains(role)) {
      return PortalShell(
          key: ValueKey('portal-$role'), role: role, modeLabel: modeLabel);
    }
    return AppShell(modeLabel: modeLabel);
  }
}
