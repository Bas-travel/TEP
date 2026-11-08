import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/theme.dart';
import '../data/api_client.dart';
import '../data/models.dart';
import 'portal/portal_kit.dart';
import 'widgets/common.dart';

/// Traveler side of the partner lifecycle: apply, see the decision, re-apply after a rejection.
/// When an admin approves the application the role becomes `partner`; "refresh" then switches the app to the partner portal.
class PartnerApplyScreen extends StatefulWidget {
  const PartnerApplyScreen({super.key});
  @override
  State<PartnerApplyScreen> createState() => _PartnerApplyScreenState();
}

class _PartnerApplyScreenState extends State<PartnerApplyScreen> {
  final _name = TextEditingController();
  final _phone = TextEditingController();
  final _desc = TextEditingController();
  int _version = 0;
  bool _busy = false;

  @override
  void dispose() {
    _name.dispose();
    _phone.dispose();
    _desc.dispose();
    super.dispose();
  }

  /// Current application, or null when the traveler has not applied yet (404).
  Future<Json?> _load() async {
    final api = AppScope.apiOf(context);
    try {
      return await api.portal('GET', '/v1/partners/me');
    } on ApiException catch (e) {
      if (e.status == 404) return null;
      rethrow;
    }
  }

  Future<void> _apply() async {
    final api = AppScope.apiOf(context);
    setState(() => _busy = true);
    final ok = await act(
      context,
      () => api.portal('POST', '/v1/partners/apply', body: {'businessName': _name.text.trim(), 'contactPhone': _phone.text.trim(), 'description': _desc.text.trim()}),
      ok: 'ส่งใบสมัครแล้ว',
    );
    if (!mounted) return;
    setState(() {
      _busy = false;
      if (ok) _version++;
    });
  }

  Widget _form({String? note}) => ListView(padding: const EdgeInsets.all(16), children: [
        const Text('เปิดร้าน/บริการของคุณบน Sabai jai Travel — ทีมงานจะตรวจสอบและแจ้งผลในแอป', style: TextStyle(color: TepColors.muted)),
        if (note != null && note.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text('เหตุผลที่ไม่อนุมัติ: $note', style: const TextStyle(color: TepColors.danger)),
        ],
        const SizedBox(height: 12),
        TextField(key: const Key('apply-name'), controller: _name, maxLength: 120, decoration: const InputDecoration(labelText: 'ชื่อธุรกิจ')),
        const SizedBox(height: 8),
        TextField(key: const Key('apply-phone'), controller: _phone, keyboardType: TextInputType.phone, maxLength: 30, decoration: const InputDecoration(labelText: 'เบอร์ติดต่อ')),
        const SizedBox(height: 8),
        TextField(key: const Key('apply-desc'), controller: _desc, maxLines: 4, maxLength: 1000, decoration: const InputDecoration(labelText: 'รายละเอียดธุรกิจ')),
        const SizedBox(height: 12),
        FilledButton(key: const Key('apply-submit'), onPressed: _busy ? null : _apply, child: const Text('ส่งใบสมัคร')),
      ]);

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('สมัครเป็นพาร์ทเนอร์')),
        body: AsyncBody<Json?>(
          key: ValueKey('apply-$_version'),
          load: _load,
          builder: (context, p, reload) {
            if (p == null) return _form();
            final st = str(p['status']);
            if (st == 'rejected') return _form(note: str(p['decisionNote']));
            return ListView(padding: const EdgeInsets.all(16), children: [
              TepCard(
                child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                  Row(children: [
                    Expanded(child: Text(str(p['businessName']), style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16))),
                    StatePill(st, partnerStatusLabels),
                  ]),
                  if (str(p['description']).isNotEmpty) Text(str(p['description'])),
                  const SizedBox(height: 8),
                  Text(
                    switch (st) {
                      'pending' => 'ใบสมัครอยู่ระหว่างการตรวจสอบ',
                      'approved' => 'อนุมัติแล้ว — กด "เข้าสู่พอร์ทัลพาร์ทเนอร์" เพื่อเริ่มจัดการสินค้าและการจอง',
                      _ => 'บัญชีพาร์ทเนอร์ถูกระงับ ${str(p['decisionNote'])}',
                    },
                    style: const TextStyle(color: TepColors.muted),
                  ),
                ]),
              ),
              const SizedBox(height: 12),
              if (st == 'approved')
                FilledButton(
                  key: const Key('apply-enter-portal'),
                  onPressed: () async {
                    final session = AppScope.of(context);
                    final nav = Navigator.of(context);
                    await guarded(context, session.refreshProfile);
                    if (nav.mounted) nav.popUntil((r) => r.isFirst);
                  },
                  child: const Text('เข้าสู่พอร์ทัลพาร์ทเนอร์'),
                )
              else
                OutlinedButton(key: const Key('apply-refresh'), onPressed: reload, child: const Text('รีเฟรชสถานะ')),
            ]);
          },
        ),
      );
}
