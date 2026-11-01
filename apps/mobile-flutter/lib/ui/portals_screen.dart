import 'package:flutter/material.dart';

import '../core/theme.dart';
import 'widgets/common.dart';

class _Kpi {
  const _Kpi(this.label, this.value, this.icon);
  final String label, value;
  final IconData icon;
}

class _Item {
  const _Item(this.icon, this.title, this.subtitle);
  final IconData icon;
  final String title, subtitle;
}

class _Portal {
  const _Portal(this.id, this.title, this.color, this.kpis, this.items);
  final String id, title;
  final Color color;
  final List<_Kpi> kpis;
  final List<_Item> items;
}

const _staffColor = Color(0xFFF59E0B);
const _adminColor = Color(0xFF7C7FC9);

const _portals = <_Portal>[
  _Portal('partner', 'พาร์ทเนอร์', TepColors.partner, [
    _Kpi('การจองวันนี้', '24', Icons.receipt_long_rounded),
    _Kpi('รายได้เดือนนี้', '฿186,400', Icons.payments_rounded),
    _Kpi('คะแนนเฉลี่ย', '4.7', Icons.star_rounded),
    _Kpi('รอยืนยัน', '5', Icons.hourglass_bottom_rounded),
  ], [
    _Item(Icons.dashboard_rounded, 'Dashboard', 'ภาพรวมยอดขายและการจอง'),
    _Item(Icons.inventory_2_rounded, 'จัดการสินค้า/บริการ', 'เพิ่ม แก้ไข เปิด-ปิดการขาย'),
    _Item(Icons.fact_check_rounded, 'การจอง', 'ยืนยัน/ปฏิเสธคำสั่งจอง'),
    _Item(Icons.account_balance_rounded, 'รายได้ & คอมมิชชัน', 'รอบการจ่ายเงินและรายงาน'),
  ]),
  _Portal('admin', 'แอดมิน', _adminColor, [
    _Kpi('ผู้ใช้ทั้งหมด', '12,480', Icons.people_rounded),
    _Kpi('พาร์ทเนอร์', '236', Icons.handshake_rounded),
    _Kpi('การจอง 24 ชม.', '1,092', Icons.receipt_long_rounded),
    _Kpi('ระบบ', 'ปกติ', Icons.monitor_heart_rounded),
  ], [
    _Item(Icons.dashboard_rounded, 'Dashboard', 'ภาพรวมแพลตฟอร์ม'),
    _Item(Icons.manage_accounts_rounded, 'จัดการผู้ใช้', 'ค้นหา ระงับ กู้คืนบัญชี'),
    _Item(Icons.store_rounded, 'จัดการพาร์ทเนอร์', 'อนุมัติและตรวจสอบพาร์ทเนอร์'),
    _Item(Icons.settings_suggest_rounded, 'ระบบ', 'ตั้งค่า บันทึกการใช้งาน (audit log)'),
  ]),
  _Portal('staff', 'พนักงาน', _staffColor, [
    _Kpi('งานค้าง', '12', Icons.assignment_late_rounded),
    _Kpi('ตั๋วช่วยเหลือ', '8', Icons.support_agent_rounded),
    _Kpi('ตรวจแล้ววันนี้', '41', Icons.verified_rounded),
    _Kpi('SLA ตอบกลับ', '2.1 ชม.', Icons.timer_rounded),
  ], [
    _Item(Icons.dashboard_rounded, 'Dashboard', 'งานของฉันวันนี้'),
    _Item(Icons.fact_check_rounded, 'ตรวจสอบการจอง', 'ตรวจสถานะและเอกสาร'),
    _Item(Icons.headset_mic_rounded, 'ช่วยเหลือลูกค้า', 'ตอบตั๋วซัพพอร์ต'),
    _Item(Icons.summarize_rounded, 'รายงาน', 'สรุปงานรายวัน'),
  ]),
];

/// Board 2.x / 3.x / 4.x — Partner, Admin and Staff portals.
/// UI preview with static sample data: these portals have no API/ERD yet (docs/requirements/mvp-scope.md).
class PortalsScreen extends StatelessWidget {
  const PortalsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: _portals.length,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('พอร์ทัล (ตัวอย่าง UI)'),
          bottom: TabBar(
            labelColor: Colors.white,
            unselectedLabelColor: Colors.white70,
            indicatorColor: Colors.white,
            tabs: [for (final p in _portals) Tab(key: Key('portal-tab-${p.id}'), text: p.title)],
          ),
        ),
        body: TabBarView(children: [for (final p in _portals) _PortalTab(p)]),
      ),
    );
  }
}

class _PortalTab extends StatelessWidget {
  const _PortalTab(this.p);
  final _Portal p;

  @override
  Widget build(BuildContext context) {
    return ListView(padding: const EdgeInsets.all(16), children: [
      Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(color: p.color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(12)),
        child: const Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Icon(Icons.info_outline_rounded, size: 18),
          SizedBox(width: 8),
          Expanded(child: Text('ตัวอย่างหน้าจอ (ข้อมูลสมมติ) — ยังไม่เชื่อม API และอยู่นอกขอบเขต MVP รอบนี้ ต้องมีบทบาท/RBAC และ /admin/v1 ก่อน', style: TextStyle(fontSize: 12))),
        ]),
      ),
      const SizedBox(height: 12),
      GridView.count(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        crossAxisCount: 2,
        mainAxisSpacing: 10,
        crossAxisSpacing: 10,
        childAspectRatio: 1.6,
        children: [
          for (final k in p.kpis)
            TepCard(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, mainAxisAlignment: MainAxisAlignment.center, children: [
                Icon(k.icon, color: p.color),
                const SizedBox(height: 4),
                Text(k.value, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800)),
                Text(k.label, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 12)),
              ]),
            ),
        ],
      ),
      const SizedBox(height: 12),
      TepCard(
        padding: EdgeInsets.zero,
        child: Column(children: [
          for (var i = 0; i < p.items.length; i++) ...[
            if (i > 0) const Divider(height: 1),
            ListTile(
              leading: Icon(p.items[i].icon, color: p.color),
              title: Text(p.items[i].title),
              subtitle: Text(p.items[i].subtitle),
            ),
          ],
        ]),
      ),
    ]);
  }
}
