import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/theme.dart';
import 'widgets/common.dart';

/// Board 1.0 / splash + auth. Register is a toggle on the same screen.
class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key, this.modeLabel = 'mock'});
  final String modeLabel;
  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _email = TextEditingController(text: 'demo@tep.dev');
  final _password = TextEditingController(text: 'Passw0rd!');
  final _first = TextEditingController();
  final _last = TextEditingController();
  bool _register = false, _busy = false, _obscure = true;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    _first.dispose();
    _last.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final session = AppScope.of(context);
    setState(() => _busy = true);
    await guarded(context, () => _register
        ? session.register(email: _email.text.trim(), password: _password.text, firstName: _first.text.trim(), lastName: _last.text.trim())
        : session.login(_email.text.trim(), _password.text));
    if (mounted) setState(() => _busy = false);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(begin: Alignment.topCenter, end: Alignment.bottomCenter, colors: [TepColors.navy, TepColors.navyBar, TepColors.partner]),
        ),
        child: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 420),
                child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
                  const Icon(Icons.travel_explore_rounded, color: Colors.white, size: 56),
                  const SizedBox(height: 12),
                  const Text('TRAVEL SUPER APP',
                      textAlign: TextAlign.center, style: TextStyle(color: Colors.white, fontSize: 26, fontWeight: FontWeight.w800, letterSpacing: 1.2)),
                  const Text('แอปเดียวเที่ยวครบ จบทุกการเดินทาง',
                      textAlign: TextAlign.center, style: TextStyle(color: Colors.white70, fontSize: 14)),
                  const SizedBox(height: 28),
                  TepCard(
                    padding: const EdgeInsets.all(20),
                    child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
                      Text(_register ? 'สร้างบัญชีใหม่' : 'เข้าสู่ระบบ', style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w700)),
                      const SizedBox(height: 16),
                      if (_register) ...[
                        Row(children: [
                          Expanded(child: TextField(key: const Key('firstName'), controller: _first, decoration: const InputDecoration(labelText: 'ชื่อ'))),
                          const SizedBox(width: 10),
                          Expanded(child: TextField(key: const Key('lastName'), controller: _last, decoration: const InputDecoration(labelText: 'นามสกุล'))),
                        ]),
                        const SizedBox(height: 12),
                      ],
                      TextField(
                        key: const Key('email'),
                        controller: _email,
                        keyboardType: TextInputType.emailAddress,
                        autofillHints: const [AutofillHints.email],
                        decoration: const InputDecoration(labelText: 'อีเมล', prefixIcon: Icon(Icons.mail_outline_rounded)),
                      ),
                      const SizedBox(height: 12),
                      TextField(
                        key: const Key('password'),
                        controller: _password,
                        obscureText: _obscure,
                        decoration: InputDecoration(
                          labelText: 'รหัสผ่าน',
                          prefixIcon: const Icon(Icons.lock_outline_rounded),
                          suffixIcon: IconButton(
                            icon: Icon(_obscure ? Icons.visibility_off_outlined : Icons.visibility_outlined),
                            onPressed: () => setState(() => _obscure = !_obscure),
                          ),
                        ),
                        onSubmitted: (_) => _busy ? null : _submit(),
                      ),
                      const SizedBox(height: 20),
                      FilledButton(
                        key: const Key('submit'),
                        onPressed: _busy ? null : _submit,
                        child: _busy
                            ? const SizedBox(height: 22, width: 22, child: CircularProgressIndicator(strokeWidth: 2.5, color: Colors.white))
                            : Text(_register ? 'ลงทะเบียน' : 'เข้าสู่ระบบ'),
                      ),
                      TextButton(
                        onPressed: _busy ? null : () => setState(() => _register = !_register),
                        child: Text(_register ? 'มีบัญชีแล้ว? เข้าสู่ระบบ' : 'ยังไม่มีบัญชี? ลงทะเบียน'),
                      ),
                    ]),
                  ),
                  const SizedBox(height: 16),
                  Text('โหมด: ${widget.modeLabel}  ·  บัญชีทดลอง demo@tep.dev', textAlign: TextAlign.center, style: const TextStyle(color: Colors.white60, fontSize: 12)),
                ]),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
