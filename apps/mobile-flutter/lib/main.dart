import 'package:flutter/material.dart';

import 'app_scope.dart';
import 'core/theme.dart';
import 'data/api_client.dart';
import 'data/http_api_client.dart';
import 'data/mock_api_client.dart';
import 'ui/login_screen.dart';
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
        title: 'TEP Travel Super App',
        debugShowCheckedModeBanner: false,
        theme: TepTheme.light(),
        home: ListenableBuilder(
          listenable: _session,
          builder: (context, _) => _session.signedIn ? AppShell(modeLabel: widget.modeLabel) : LoginScreen(modeLabel: widget.modeLabel),
        ),
      ),
    );
  }
}
