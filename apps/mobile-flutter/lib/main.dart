import 'package:flutter/material.dart';

import 'app_scope.dart';
import 'core/theme.dart';
import 'data/api_client.dart';
import 'data/http_api_client.dart';
import 'data/mock_api_client.dart';
import 'ui/login_screen.dart';
import 'ui/portal/portal_shell.dart';
import 'ui/shell.dart';

/// `flutter run --dart-define=API_MODE=http --dart-define=API_BASE_URL=http://10.0.2.2:8080`
const _apiMode = String.fromEnvironment('API_MODE', defaultValue: 'mock');
const _apiBaseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: 'http://10.0.2.2:8080');

void main() {
  final ApiClient api = _apiMode == 'http' ? HttpApiClient(baseUrl: _apiBaseUrl) : MockApiClient();
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
        title: 'Sabai jai Travel',
        debugShowCheckedModeBanner: false,
        theme: TepTheme.light(),
        home: ListenableBuilder(
          listenable: _session,
          builder: (context, _) {
            final user = _session.user;
            if (user == null) return LoginScreen(modeLabel: widget.modeLabel);
            // Back-office roles get their own shell; everyone else (traveler, moderator) uses the traveler app.
            if (const {'partner', 'admin', 'staff'}.contains(user.role)) {
              return PortalShell(key: ValueKey('portal-${user.role}'), role: user.role, modeLabel: widget.modeLabel);
            }
            return AppShell(modeLabel: widget.modeLabel);
          },
        ),
      ),
    );
  }
}
