import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/responsive.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'widgets/brand.dart';
import 'widgets/common.dart';

class RoleShell extends StatefulWidget {
  const RoleShell({super.key, required this.role, this.modeLabel = 'mock'});
  final String role;
  final String modeLabel;

  @override
  State<RoleShell> createState() => _RoleShellState();
}

class _RoleShellState extends State<RoleShell> {
  int _index = 0;

  String get _title {
    return switch (widget.role) {
      AccessRole.partner => 'Partner Workspace',
      AccessRole.staff => 'Staff Workspace',
      AccessRole.moderator => 'Moderation Center',
      AccessRole.admin => 'Admin Console',
      _ => 'Traveler Workspace',
    };
  }

  List<_RoleNavItem> get _navigation {
    return switch (widget.role) {
      AccessRole.partner => const [
          _RoleNavItem('ภาพรวม', Icons.dashboard_rounded),
          _RoleNavItem('สินค้าและบริการ', Icons.storefront_rounded),
          _RoleNavItem('การจอง', Icons.receipt_long_rounded),
          _RoleNavItem('รายได้', Icons.payments_rounded),
          _RoleNavItem('บัญชี', Icons.person_rounded),
        ],
      AccessRole.admin => const [
          _RoleNavItem('ภาพรวมระบบ', Icons.dashboard_rounded),
          _RoleNavItem('ผู้ใช้', Icons.people_alt_rounded),
          _RoleNavItem('พาร์ทเนอร์', Icons.handshake_rounded),
          _RoleNavItem('รายงาน', Icons.analytics_rounded),
          _RoleNavItem('ตั้งค่า', Icons.settings_rounded),
        ],
      AccessRole.staff => const [
          _RoleNavItem('งานวันนี้', Icons.assignment_rounded),
          _RoleNavItem('ตรวจการจอง', Icons.fact_check_rounded),
          _RoleNavItem('ตั๋วช่วยเหลือ', Icons.support_agent_rounded),
          _RoleNavItem('รายงาน', Icons.summarize_rounded),
          _RoleNavItem('บัญชี', Icons.person_rounded),
        ],
      AccessRole.moderator => const [
          _RoleNavItem('ภาพรวม', Icons.dashboard_rounded),
          _RoleNavItem('รายงานเนื้อหา', Icons.report_rounded),
          _RoleNavItem('โพสต์', Icons.forum_rounded),
          _RoleNavItem('สมาชิก', Icons.group_rounded),
          _RoleNavItem('บันทึกการดำเนินการ', Icons.history_rounded),
        ],
      _ => const [_RoleNavItem('หน้าแรก', Icons.home_rounded)],
    };
  }

  @override
  Widget build(BuildContext context) {
    final user = AppScope.of(context).user;
    final nav = _navigation;
    if (_index >= nav.length) _index = 0;
    final extended =
        MediaQuery.sizeOf(context).width >= TepBreakpoints.expanded;
    final compact = MediaQuery.sizeOf(context).width < TepBreakpoints.compact;

    return Scaffold(
      backgroundColor: TepColors.background,
      body: Row(children: [
        if (!compact)
          NavigationRail(
            key: const Key('role-navigation'),
            extended: extended,
            minExtendedWidth: 230,
            selectedIndex: _index,
            onDestinationSelected: (value) => setState(() => _index = value),
            leading: Padding(
              padding: const EdgeInsets.fromLTRB(12, 16, 12, 16),
              child: extended
                  ? const TepBrandMark(compact: true)
                  : const CircleAvatar(
                      backgroundColor: TepColors.primaryLight,
                      child: Text('TEP',
                          style: TextStyle(
                              color: TepColors.primaryDeep,
                              fontWeight: FontWeight.w900,
                              fontSize: 11))),
            ),
            trailing: Expanded(
                child: Align(
                    alignment: Alignment.bottomCenter,
                    child: Padding(
                        padding: const EdgeInsets.only(bottom: 16),
                        child: IconButton(
                            key: const Key('role-logout'),
                            tooltip: 'ออกจากระบบ',
                            onPressed: () =>
                                guarded(context, AppScope.of(context).logout),
                            icon: const Icon(Icons.logout_rounded,
                                color: TepColors.danger))))),
            destinations: [
              for (final item in nav)
                NavigationRailDestination(
                    icon: Icon(item.icon, color: TepColors.muted),
                    selectedIcon: Icon(item.icon, color: TepColors.primary),
                    label: Text(item.label))
            ],
            backgroundColor: Colors.white,
          ),
        Expanded(
          child: Column(children: [
            Container(
              color: Colors.white,
              child: Column(children: [
                TepPageHeader(
                  title: _title,
                  subtitle:
                      '${user?.fullName ?? ''} · ${widget.modeLabel.toUpperCase()}',
                ),
                Align(
                  alignment: Alignment.centerRight,
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
                    child: Chip(
                      label: Text(user?.role ?? widget.role),
                      avatar: const Icon(Icons.verified_user_outlined,
                          size: 16, color: TepColors.primary),
                    ),
                  ),
                ),
              ]),
            ),
            const Divider(height: 1),
            Expanded(
                child: _RoleDashboard(
                    role: widget.role, section: nav[_index].label)),
          ]),
        ),
      ]),
      bottomNavigationBar: compact
          ? NavigationBar(
              selectedIndex: _index,
              onDestinationSelected: (value) => setState(() => _index = value),
              destinations: [
                  for (final item in nav.take(5))
                    NavigationDestination(
                        icon: Icon(item.icon), label: item.label)
                ])
          : null,
    );
  }
}

class _RoleNavItem {
  const _RoleNavItem(this.label, this.icon);
  final String label;
  final IconData icon;
}

class _RoleDashboard extends StatelessWidget {
  const _RoleDashboard({required this.role, required this.section});
  final String role;
  final String section;

  @override
  Widget build(BuildContext context) {
    final roleData = switch (role) {
      AccessRole.partner => const [
          ('การจองวันนี้', '24', Icons.receipt_long_rounded),
          ('รายได้เดือนนี้', '฿186,400', Icons.payments_rounded),
          ('คะแนนเฉลี่ย', '4.7', Icons.star_rounded),
          ('รอยืนยัน', '5', Icons.pending_actions_rounded)
        ],
      AccessRole.admin => const [
          ('ผู้ใช้ทั้งหมด', '12,480', Icons.people_alt_rounded),
          ('พาร์ทเนอร์', '236', Icons.handshake_rounded),
          ('การจอง 24 ชม.', '1,092', Icons.receipt_long_rounded),
          ('สถานะระบบ', 'ปกติ', Icons.health_and_safety_rounded)
        ],
      AccessRole.staff => const [
          ('งานค้าง', '12', Icons.assignment_late_rounded),
          ('ตั๋วช่วยเหลือ', '8', Icons.support_agent_rounded),
          ('ตรวจแล้ววันนี้', '41', Icons.fact_check_rounded),
          ('SLA ตอบกลับ', '2.1 ชม.', Icons.timer_rounded)
        ],
      AccessRole.moderator => const [
          ('รายงานใหม่', '18', Icons.report_rounded),
          ('โพสต์ถูกซ่อน', '7', Icons.visibility_off_rounded),
          ('รอตรวจ', '5', Icons.pending_actions_rounded),
          ('ดำเนินการวันนี้', '31', Icons.gavel_rounded)
        ],
      _ => const [
          ('การจอง', '3', Icons.receipt_long_rounded),
          ('ทริปบันทึกไว้', '6', Icons.bookmark_rounded),
          ('ยอด Wallet', '฿1,500', Icons.account_balance_wallet_rounded),
          ('คะแนนสะสม', '1,240', Icons.workspace_premium_rounded)
        ],
    };

    return TepConstrainedContent(
        maxWidth: 1400,
        child:
            ListView(padding: TepBreakpoints.pagePadding(context), children: [
          const SizedBox(height: 16),
          Text(section,
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.w800, color: TepColors.ink)),
          const SizedBox(height: 12),
          LayoutBuilder(builder: (context, constraints) {
            final columns = TepBreakpoints.gridColumns(context,
                compact: 2, medium: 2, expanded: 4);
            return GridView.count(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                crossAxisCount: columns,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: constraints.maxWidth < 700 ? 1.35 : 1.8,
                children: [
                  for (final item in roleData)
                    Card(
                        color: Colors.white,
                        child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(item.$3, color: TepColors.primary),
                                  const SizedBox(height: 8),
                                  Text(item.$2,
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: const TextStyle(
                                          fontSize: 22,
                                          fontWeight: FontWeight.w900,
                                          color: TepColors.ink)),
                                  Text(item.$1,
                                      style: const TextStyle(
                                          color: TepColors.muted))
                                ])))
                ]);
          }),
          const SizedBox(height: 16),
          Card(
              child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('รายการล่าสุด',
                            style: TextStyle(
                                fontSize: 18, fontWeight: FontWeight.w800)),
                        const SizedBox(height: 12),
                        for (final label in [
                          'มีรายการใหม่รอตรวจสอบ',
                          'อัปเดตสถานะล่าสุด 5 นาทีที่แล้ว',
                          'รายงานพร้อมส่งต่อ'
                        ]) ...[
                          ListTile(
                              leading: const Icon(Icons.circle_outlined,
                                  color: TepColors.primary),
                              title: Text(label),
                              trailing:
                                  const Icon(Icons.chevron_right_rounded)),
                          const Divider(height: 1)
                        ]
                      ]))),
        ]));
  }
}
