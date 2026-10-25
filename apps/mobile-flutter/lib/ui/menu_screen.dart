import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'account_screens.dart';
import 'city_guide_screen.dart';
import 'community_screen.dart';
import 'portals_screen.dart';
import 'reviews_screen.dart';
import 'support_screen.dart';
import 'trips_screen.dart';
import 'widgets/common.dart';

class _Entry {
  const _Entry(this.id, this.icon, this.label, {this.tab, this.page});
  final String id, label;
  final IconData icon;
  final int? tab; // bottom-nav tab to switch to
  final Widget Function()? page;
}

/// Board "Main menu" — all 12 entries, plus the extra traveler features from the board.
class MenuScreen extends StatelessWidget {
  const MenuScreen({super.key, required this.onTab});

  /// Switches the bottom-navigation tab of the shell (the menu pops itself first).
  final void Function(int tab) onTab;

  static final _main = <_Entry>[
    const _Entry('home', Icons.home_rounded, 'หน้าแรก (Home)', tab: 0),
    const _Entry('search', Icons.search_rounded, 'ค้นหา (Search)', tab: 1),
    const _Entry('bookings', Icons.receipt_long_rounded, 'การจองของฉัน (My Bookings)', tab: 2),
    _Entry('trips', Icons.luggage_rounded, 'ทริปของฉัน (My Trips)', page: () => const TripsScreen()),
    _Entry('tickets', Icons.confirmation_number_rounded, 'ตั๋ว / พาส (Tickets & Pass)', page: () => const TicketsScreen()),
    const _Entry('wallet', Icons.account_balance_wallet_rounded, 'กระเป๋าเงิน (Wallet)', tab: 3),
    _Entry('coupons', Icons.local_offer_rounded, 'คูปอง (Coupons)', page: () => const CouponsScreen()),
    _Entry('notifications', Icons.notifications_rounded, 'แจ้งเตือน (Notifications)', page: () => const NotificationsScreen()),
    _Entry('messages', Icons.chat_bubble_rounded, 'ข้อความ (Messages)', page: () => const MessagesScreen()),
    const _Entry('profile', Icons.person_rounded, 'โปรไฟล์ (Profile)', tab: 4),
    _Entry('settings', Icons.settings_rounded, 'ตั้งค่า (Settings)', page: () => const SettingsScreen()),
    _Entry('support', Icons.help_rounded, 'ช่วยเหลือ (Support)', page: () => const SupportScreen()),
  ];

  static final _more = <_Entry>[
    _Entry('ai', Icons.auto_awesome_rounded, 'AI ผู้ช่วยวางแผน', page: () => const AiPlannerScreen()),
    _Entry('community', Icons.forum_rounded, 'ชุมชนนักเดินทาง', page: () => const CommunityScreen()),
    _Entry('guide', Icons.explore_rounded, 'ไกด์เมือง', page: () => const CityGuideListScreen()),
    _Entry('reviews', Icons.star_rounded, 'รีวิวของฉัน', page: () => const ReviewsScreen()),
    _Entry('map', Icons.map_rounded, 'แผนที่ใกล้ฉัน', page: () => const ComingSoonScreen(title: 'แผนที่ใกล้ฉัน', icon: Icons.map_rounded, note: 'MVP รอบนี้เปิดแผนที่ผ่านลิงก์ภายนอก (deep link) เท่านั้น — แผนที่ในแอปอยู่ในแผนเฟสถัดไป')),
    _Entry('weather', Icons.wb_sunny_rounded, 'พยากรณ์อากาศ', page: () => const ComingSoonScreen(title: 'พยากรณ์อากาศ', icon: Icons.wb_sunny_rounded, note: 'ต้องเชื่อมผู้ให้บริการสภาพอากาศภายนอก — อยู่นอกขอบเขต MVP รอบนี้')),
    _Entry('portals', Icons.dashboard_customize_rounded, 'พอร์ทัลพาร์ทเนอร์ / แอดมิน / พนักงาน (ตัวอย่าง)', page: () => const PortalsScreen()),
  ];

  void _open(BuildContext context, _Entry e) {
    final nav = Navigator.of(context);
    if (e.tab != null) {
      nav.pop();
      onTab(e.tab!);
    } else {
      nav.push(MaterialPageRoute<void>(builder: (_) => e.page!()));
    }
  }

  @override
  Widget build(BuildContext context) {
    Widget group(String title, List<_Entry> entries) => Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Padding(padding: const EdgeInsets.fromLTRB(4, 16, 4, 8), child: Text(title, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16))),
          TepCard(
            padding: EdgeInsets.zero,
            child: Column(children: [
              for (var i = 0; i < entries.length; i++) ...[
                if (i > 0) const Divider(height: 1),
                ListTile(
                  key: Key('menu-${entries[i].id}'),
                  leading: Icon(entries[i].icon, color: TepColors.primary),
                  title: Text(entries[i].label),
                  trailing: const Icon(Icons.chevron_right_rounded),
                  onTap: () => _open(context, entries[i]),
                ),
              ],
            ]),
          ),
        ]);

    return Scaffold(
      appBar: AppBar(title: const Text('เมนูหลัก')),
      body: ListView(padding: const EdgeInsets.fromLTRB(16, 0, 16, 24), children: [
        group('เมนูหลัก', _main),
        group('ฟีเจอร์เพิ่มเติม', _more),
      ]),
    );
  }
}

/// Board 1.16 — tickets & pass: confirmed bookings rendered as e-tickets.
class TicketsScreen extends StatelessWidget {
  const TicketsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('ตั๋ว / พาส')),
      body: AsyncBody<List<BookingSummary>>(
        load: () => api.bookings(status: 'confirmed'),
        builder: (context, items, reload) {
          if (items.isEmpty) return const EmptyState(Icons.confirmation_number_outlined, 'ยังไม่มีตั๋วที่ใช้งานได้\nตั๋วจะแสดงหลังการจองได้รับการยืนยัน');
          return RefreshIndicator(
            onRefresh: reload,
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: 12),
              itemBuilder: (context, i) {
                final b = items[i];
                return TepCard(
                  key: Key('ticket-${b.id}'),
                  child: Row(children: [
                    Expanded(
                      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                        Text(b.title, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16)),
                        const SizedBox(height: 4),
                        Text(b.bookingNumber, style: const TextStyle(color: TepColors.muted)),
                        Text(dmy(b.date), style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                        const SizedBox(height: 6),
                        StatusBadge(b.status),
                      ]),
                    ),
                    const SizedBox(width: 12),
                    const Icon(Icons.qr_code_2_rounded, size: 72, color: TepColors.ink),
                  ]),
                );
              },
            ),
          );
        },
      ),
    );
  }
}

/// Board 1.22 — messages. Live chat needs a WebSocket service (deferred); this is a read-only inbox stub.
class MessagesScreen extends StatelessWidget {
  const MessagesScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('ข้อความ')),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        const TepCard(
          padding: EdgeInsets.zero,
          child: ListTile(
            leading: CircleAvatar(backgroundColor: TepColors.primary, child: Icon(Icons.support_agent_rounded, color: Colors.white)),
            title: Text('ทีมงาน TEP', style: TextStyle(fontWeight: FontWeight.w700)),
            subtitle: Text('ยินดีต้อนรับ! หากมีคำถามเกี่ยวกับการจอง ส่งข้อความผ่านหน้า ติดต่อเรา ได้เลย'),
          ),
        ),
        const SizedBox(height: 12),
        Text('แชทสดกับพาร์ทเนอร์และทีมงานจะเปิดใช้งานเมื่อมีบริการ WebSocket (อยู่นอกขอบเขต MVP รอบนี้)', style: TextStyle(color: TepColors.muted.withValues(alpha: 0.9), fontSize: 12)),
        const SizedBox(height: 12),
        OutlinedButton.icon(
          key: const Key('messages-contact'),
          onPressed: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const SupportScreen(initialTab: 1))),
          icon: const Icon(Icons.mail_outline_rounded),
          label: const Text('ติดต่อเรา'),
        ),
      ]),
    );
  }
}

class ComingSoonScreen extends StatelessWidget {
  const ComingSoonScreen({super.key, required this.title, required this.icon, required this.note});
  final String title, note;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            Icon(icon, size: 64, color: TepColors.line),
            const SizedBox(height: 16),
            const Text('เร็ว ๆ นี้', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w800)),
            const SizedBox(height: 8),
            Text(note, textAlign: TextAlign.center, style: const TextStyle(color: TepColors.muted)),
          ]),
        ),
      ),
    );
  }
}
