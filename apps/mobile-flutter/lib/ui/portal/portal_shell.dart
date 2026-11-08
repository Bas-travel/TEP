import 'package:flutter/material.dart';

import '../../app_scope.dart';
import 'admin_pages.dart';
import 'partner_pages.dart';
import 'portal_kit.dart';
import 'staff_pages.dart';

class _Dest {
  const _Dest(this.label, this.icon, this.page);
  final String label;
  final IconData icon;
  final Widget Function() page;
}

List<_Dest> _destinationsFor(String role) => switch (role) {
      'partner' => [
          _Dest('ภาพรวม', Icons.dashboard_rounded, () => const PartnerDashboardPage()),
          _Dest('สินค้า/บริการ', Icons.inventory_2_rounded, () => const PartnerListingsPage()),
          _Dest('การจอง', Icons.fact_check_rounded, () => const PartnerBookingsPage()),
          _Dest('รายได้', Icons.account_balance_rounded, () => const PartnerEarningsPage()),
        ],
      'admin' => [
          _Dest('ภาพรวม', Icons.dashboard_rounded, () => const AdminDashboardPage()),
          _Dest('ผู้ใช้', Icons.manage_accounts_rounded, () => const AdminUsersPage()),
          _Dest('พาร์ทเนอร์', Icons.store_rounded, () => const AdminPartnersPage()),
          _Dest('ระบบ', Icons.settings_suggest_rounded, () => const AdminSystemPage()),
        ],
      _ => [
          _Dest('ภาพรวม', Icons.dashboard_rounded, () => const StaffDashboardPage()),
          _Dest('ตรวจการจอง', Icons.fact_check_rounded, () => const StaffBookingsPage()),
          _Dest('ช่วยเหลือ', Icons.headset_mic_rounded, () => const StaffTicketsPage()),
          _Dest('รายงาน', Icons.summarize_rounded, () => const StaffReportPage()),
        ],
    };

/// Back-office shell shown instead of the traveler app for the roles partner, admin and staff (boards 2.x / 3.x / 4.x).
/// Each role gets its own bottom menu; every page talks to its own API prefix, which the server guards by role.
class PortalShell extends StatefulWidget {
  const PortalShell({super.key, required this.role, this.modeLabel = 'mock'});
  final String role, modeLabel;
  @override
  State<PortalShell> createState() => _PortalShellState();
}

class _PortalShellState extends State<PortalShell> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    final dests = _destinationsFor(widget.role);
    final session = AppScope.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Column(mainAxisSize: MainAxisSize.min, children: [
          Text('${roleLabels[widget.role] ?? widget.role} · ${dests[_tab].label}', key: const Key('portal-title'), style: const TextStyle(fontSize: 17, fontWeight: FontWeight.w700)),
          Text('${session.user?.fullName ?? ''} · ${widget.modeLabel}', style: const TextStyle(fontSize: 11, color: Colors.white70)),
        ]),
        actions: [
          IconButton(key: const Key('portal-logout'), tooltip: 'ออกจากระบบ', icon: const Icon(Icons.logout_rounded), onPressed: session.logout),
        ],
      ),
      body: KeyedSubtree(key: ValueKey('${widget.role}-$_tab'), child: dests[_tab].page()),
      bottomNavigationBar: NavigationBar(
        key: const Key('portal-nav'),
        selectedIndex: _tab,
        onDestinationSelected: (i) => setState(() => _tab = i),
        destinations: [for (final d in dests) NavigationDestination(icon: Icon(d.icon), label: d.label)],
      ),
    );
  }
}
