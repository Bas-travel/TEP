import 'package:flutter/material.dart';

import 'data/api_client.dart';
import 'data/models.dart';

/// Holds the API client and the signed-in user for the whole widget tree.
class SessionController extends ChangeNotifier {
  SessionController(this.api);
  final ApiClient api;
  UserProfile? _user;
  UserProfile? get user => _user;
  bool get signedIn => _user != null;

  Future<void> login(String email, String password) async {
    await api.login(email, password);
    _user = await api.me();
    notifyListeners();
  }

  Future<void> register({required String email, required String password, required String firstName, required String lastName}) async {
    await api.register(email: email, password: password, firstName: firstName, lastName: lastName);
    _user = await api.me();
    notifyListeners();
  }

  Future<void> logout() async {
    await api.logout();
    _user = null;
    notifyListeners();
  }
}

class AppScope extends InheritedNotifier<SessionController> {
  const AppScope({super.key, required SessionController session, required super.child}) : super(notifier: session);

  static SessionController of(BuildContext context) {
    final s = context.dependOnInheritedWidgetOfExactType<AppScope>()?.notifier;
    assert(s != null, 'AppScope missing above this context');
    return s!;
  }

  static ApiClient apiOf(BuildContext context) => of(context).api;
}
