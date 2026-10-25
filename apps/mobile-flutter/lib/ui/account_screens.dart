import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models_engage.dart';
import 'widgets/common.dart';

/// Board 1.18 — coupons.
class CouponsScreen extends StatelessWidget {
  const CouponsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('คูปองของฉัน')),
      body: AsyncBody<List<Coupon>>(
        load: api.coupons,
        builder: (context, items, reload) {
          if (items.isEmpty) return const EmptyState(Icons.local_offer_outlined, 'ยังไม่มีคูปอง');
          return RefreshIndicator(
            onRefresh: reload,
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (context, i) => _CouponCard(items[i]),
            ),
          );
        },
      ),
    );
  }
}

class _CouponCard extends StatelessWidget {
  const _CouponCard(this.c);
  final Coupon c;

  @override
  Widget build(BuildContext context) {
    final color = c.usable ? TepColors.success : TepColors.muted;
    return TepCard(
      child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Container(
          width: 52,
          height: 52,
          decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(14)),
          child: Icon(Icons.local_offer_rounded, color: color),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(c.title, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
            const SizedBox(height: 2),
            Text(c.detail, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 12)),
            const SizedBox(height: 4),
            Text('หมดอายุ ${c.expiry}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
            const SizedBox(height: 8),
            Row(children: [
              Flexible(
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(border: Border.all(color: color), borderRadius: BorderRadius.circular(8)),
                  child: Text(c.code, maxLines: 1, overflow: TextOverflow.ellipsis, style: TextStyle(color: color, fontWeight: FontWeight.w800, fontSize: 12)),
                ),
              ),
              const SizedBox(width: 8),
              if (c.usable)
                TextButton(
                  key: Key('coupon-copy-${c.code}'),
                  onPressed: () async {
                    await Clipboard.setData(ClipboardData(text: c.code));
                    if (context.mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('คัดลอกรหัส ${c.code} แล้ว')));
                    }
                  },
                  child: const Text('คัดลอก'),
                )
              else
                StatusBadge(c.status == 'used' ? 'completed' : 'cancelled'),
            ]),
          ]),
        ),
      ]),
    );
  }
}

/// Board 1.21 — notifications with unread state.
class NotificationsScreen extends StatefulWidget {
  const NotificationsScreen({super.key});
  @override
  State<NotificationsScreen> createState() => _NotificationsScreenState();
}

class _NotificationsScreenState extends State<NotificationsScreen> {
  int _version = 0;

  static IconData _icon(String t) => switch (t) {
        'promo' => Icons.local_offer_rounded,
        'trip' => Icons.luggage_rounded,
        'booking' => Icons.receipt_long_rounded,
        _ => Icons.info_outline_rounded,
      };

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('การแจ้งเตือน'), actions: [
        TextButton(
          key: const Key('read-all'),
          onPressed: () async {
            await guarded(context, api.markAllNotificationsRead);
            if (mounted) setState(() => _version++);
          },
          child: const Text('อ่านทั้งหมด', style: TextStyle(color: Colors.white)),
        ),
      ]),
      body: AsyncBody<List<AppNotification>>(
        key: ValueKey('n-$_version'),
        load: api.notifications,
        builder: (context, items, reload) {
          if (items.isEmpty) return const EmptyState(Icons.notifications_none_rounded, 'ไม่มีการแจ้งเตือน');
          return RefreshIndicator(
            onRefresh: reload,
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: 8),
              itemBuilder: (context, i) {
                final n = items[i];
                return TepCard(
                  padding: EdgeInsets.zero,
                  child: ListTile(
                    key: Key('notif-${n.id}'),
                    leading: CircleAvatar(
                      backgroundColor: TepColors.navyBar.withValues(alpha: 0.10),
                      child: Icon(_icon(n.type), color: TepColors.navyBar, size: 20),
                    ),
                    title: Text(n.title, style: TextStyle(fontWeight: n.read ? FontWeight.w500 : FontWeight.w800)),
                    subtitle: Text('${n.message}\n${dmy(n.createdAt)}'),
                    isThreeLine: true,
                    trailing: n.read ? null : const Icon(Icons.circle, size: 10, color: TepColors.danger),
                    onTap: n.read
                        ? null
                        : () async {
                            await guarded(context, () => api.markNotificationRead(n.id));
                            if (mounted) setState(() => _version++);
                          },
                  ),
                );
              },
            ),
          );
        },
      ),
    );
  }
}

/// Board 1.26 — settings. Notification switches persist through `PATCH /me/settings`.
class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('ตั้งค่า')),
      body: AsyncBody<UserSettings>(
        load: api.settings,
        builder: (context, s, _) => _SettingsForm(initial: s),
      ),
    );
  }
}

class _SettingsForm extends StatefulWidget {
  const _SettingsForm({required this.initial});
  final UserSettings initial;
  @override
  State<_SettingsForm> createState() => _SettingsFormState();
}

class _SettingsFormState extends State<_SettingsForm> {
  late UserSettings _s = widget.initial;

  Future<void> _save(UserSettings next) async {
    final prev = _s;
    setState(() => _s = next);
    final api = AppScope.apiOf(context);
    final saved = await guarded(context, () => api.updateSettings(next));
    if (!mounted) return;
    if (saved == null) {
      setState(() => _s = prev);
    } else {
      setState(() => _s = saved);
    }
  }

  @override
  Widget build(BuildContext context) {
    return ListView(padding: const EdgeInsets.all(16), children: [
      const Padding(padding: EdgeInsets.only(bottom: 8), child: Text('การแจ้งเตือน', style: TextStyle(fontWeight: FontWeight.w700, fontSize: 16))),
      TepCard(
        padding: EdgeInsets.zero,
        child: Column(children: [
          SwitchListTile(key: const Key('set-push'), title: const Text('Push notification'), value: _s.push, onChanged: (v) => _save(_s.copyWith(push: v))),
          const Divider(height: 1),
          SwitchListTile(key: const Key('set-email'), title: const Text('อีเมล'), value: _s.email, onChanged: (v) => _save(_s.copyWith(email: v))),
          const Divider(height: 1),
          SwitchListTile(key: const Key('set-trip'), title: const Text('แจ้งเตือนทริปและการจอง'), value: _s.trip, onChanged: (v) => _save(_s.copyWith(trip: v))),
          const Divider(height: 1),
          SwitchListTile(key: const Key('set-promo'), title: const Text('โปรโมชันและคูปอง'), value: _s.promo, onChanged: (v) => _save(_s.copyWith(promo: v))),
        ]),
      ),
      const Padding(padding: EdgeInsets.only(top: 20, bottom: 8), child: Text('ทั่วไป', style: TextStyle(fontWeight: FontWeight.w700, fontSize: 16))),
      TepCard(
        padding: EdgeInsets.zero,
        child: Column(children: [
          ListTile(leading: const Icon(Icons.language_rounded), title: const Text('ภาษา'), trailing: Text(_s.language == 'th' ? 'ไทย' : _s.language)),
          const Divider(height: 1),
          ListTile(leading: const Icon(Icons.payments_outlined), title: const Text('สกุลเงิน'), trailing: Text(_s.currency)),
          const Divider(height: 1),
          const ListTile(leading: Icon(Icons.dark_mode_outlined), title: Text('ธีม'), subtitle: Text('โหมดมืดจะเปิดใช้ในเวอร์ชันถัดไป'), trailing: Text('สว่าง')),
        ]),
      ),
    ]);
  }
}
