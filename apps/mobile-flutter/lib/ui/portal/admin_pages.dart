import 'package:flutter/material.dart';

import '../../app_scope.dart';
import '../../core/format.dart';
import '../../core/theme.dart';
import '../../data/api_client.dart';
import '../widgets/common.dart';
import 'portal_kit.dart';

/// Board 3.1 — platform dashboard.
class AdminDashboardPage extends StatelessWidget {
  const AdminDashboardPage({super.key});
  @override
  Widget build(BuildContext context) => PortalPage(
        path: '/admin/v1/dashboard',
        builder: (context, d) {
          final partners = jmap(d['partners']);
          return [
            KpiGrid([
              Kpi('ผู้ใช้ทั้งหมด', '${numOf(d['users'])}', Icons.people_rounded),
              Kpi('ถูกระงับ', '${numOf(d['usersSuspended'])}', Icons.block_rounded),
              Kpi('พาร์ทเนอร์ที่ใช้งาน', '${numOf(partners['approved'])}', Icons.handshake_rounded),
              Kpi('รออนุมัติ', '${numOf(partners['pending'])}', Icons.hourglass_bottom_rounded),
              Kpi('รายการที่เปิดขาย', '${numOf(d['listingsPublished'])}', Icons.storefront_rounded),
              Kpi('การจองวันนี้', '${numOf(d['bookingsToday'])}', Icons.receipt_long_rounded),
              Kpi('การจองทั้งหมด', '${numOf(d['bookingsTotal'])}', Icons.inventory_rounded),
              Kpi('รายได้รวม', thb(numOf(d['revenue'])), Icons.payments_rounded),
            ]),
          ];
        },
      );
}

/// Board 3.2 — users: search, change role, suspend / reactivate.
class AdminUsersPage extends StatelessWidget {
  const AdminUsersPage({super.key});
  @override
  Widget build(BuildContext context) => PortalList(
        path: '/admin/v1/users',
        searchHint: 'ค้นหาชื่อหรืออีเมล',
        emptyText: 'ไม่พบผู้ใช้',
        filters: const [
          ListFilter('ทั้งหมด'),
          ListFilter('นักท่องเที่ยว', {'role': 'traveler'}),
          ListFilter('พาร์ทเนอร์', {'role': 'partner'}),
          ListFilter('พนักงาน', {'role': 'staff'}),
          ListFilter('แอดมิน', {'role': 'admin'}),
        ],
        itemBuilder: (context, u, reload) => TepCard(
          key: Key('user-${u['id']}'),
          onTap: () async {
            final changed = await showModalBottomSheet<bool>(
              context: context,
              isScrollControlled: true,
              builder: (_) => _UserSheet(user: u, selfId: AppScope.of(context).user?.id ?? 0),
            );
            if (changed == true) await reload();
          },
          child: Row(children: [
            const CircleAvatar(backgroundColor: TepColors.secondary, child: Icon(Icons.person_rounded, color: TepColors.ink)),
            const SizedBox(width: 12),
            Expanded(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text(str(u['name']), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                Text(str(u['email']), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                Text('สมัคร ${dateOf(u['createdAt'])}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
              ]),
            ),
            Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
              Pill(roleLabels[u['role']] ?? str(u['role']), color: TepColors.primary),
              const SizedBox(height: 4),
              StatePill(str(u['status']), userStatusLabels),
            ]),
          ]),
        ),
      );
}

class _UserSheet extends StatefulWidget {
  const _UserSheet({required this.user, required this.selfId});
  final Json user;
  final int selfId;
  @override
  State<_UserSheet> createState() => _UserSheetState();
}

class _UserSheetState extends State<_UserSheet> {
  late String _role = str(widget.user['role']);
  late bool _active = widget.user['status'] != 'suspended';
  bool _busy = false;

  bool get _isSelf => widget.user['id'] == widget.selfId;
  bool get _isPartner => widget.user['role'] == 'partner';
  bool get _deleted => widget.user['status'] == 'deleted';

  Future<void> _save() async {
    final api = AppScope.apiOf(context);
    final body = <String, Object?>{
      if (_role != widget.user['role']) 'role': _role,
      if (_active != (widget.user['status'] != 'suspended')) 'status': _active ? 'active' : 'suspended',
    };
    if (body.isEmpty) {
      Navigator.of(context).pop(false);
      return;
    }
    setState(() => _busy = true);
    final ok = await act(context, () => api.portal('PATCH', '/admin/v1/users/${widget.user['id']}', body: body), ok: 'บันทึกแล้ว');
    if (!mounted) return;
    setState(() => _busy = false);
    if (ok) Navigator.of(context).pop(true);
  }

  @override
  Widget build(BuildContext context) {
    final locked = _isSelf || _deleted;
    return Padding(
      padding: EdgeInsets.fromLTRB(16, 16, 16, 16 + MediaQuery.of(context).viewInsets.bottom),
      child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text(str(widget.user['name']), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800)),
        Text(str(widget.user['email']), style: const TextStyle(color: TepColors.muted)),
        const SizedBox(height: 16),
        if (_isSelf) const Text('ไม่สามารถเปลี่ยนบทบาทหรือสถานะของตัวเองได้', style: TextStyle(color: TepColors.muted)),
        if (_isPartner) const Text('บัญชีพาร์ทเนอร์จัดการที่เมนู "พาร์ทเนอร์" (อนุมัติ/ระงับ)', style: TextStyle(color: TepColors.muted)),
        if (!_isPartner)
          DropdownButtonFormField<String>(
            key: const Key('user-role'),
            initialValue: _role,
            decoration: const InputDecoration(labelText: 'บทบาท'),
            items: [for (final r in const ['traveler', 'staff', 'moderator', 'admin']) DropdownMenuItem(value: r, child: Text(roleLabels[r] ?? r))],
            onChanged: locked ? null : (v) => setState(() => _role = v ?? _role),
          ),
        SwitchListTile(
          key: const Key('user-active'),
          contentPadding: EdgeInsets.zero,
          title: const Text('เปิดใช้งานบัญชี'),
          subtitle: const Text('การระงับจะออกจากระบบทุกอุปกรณ์ทันที'),
          value: _active,
          onChanged: locked ? null : (v) => setState(() => _active = v),
        ),
        const SizedBox(height: 8),
        FilledButton(key: const Key('user-save'), onPressed: (_busy || locked) ? null : _save, child: const Text('บันทึก')),
      ]),
    );
  }
}

/// Board 3.3 — partners (applications) and listing moderation.
class AdminPartnersPage extends StatelessWidget {
  const AdminPartnersPage({super.key});
  @override
  Widget build(BuildContext context) => const DefaultTabController(
        length: 2,
        child: Column(children: [
          TabBar(
            labelColor: TepColors.primary,
            indicatorColor: TepColors.primary,
            tabs: [Tab(key: Key('tab-partners'), text: 'พาร์ทเนอร์'), Tab(key: Key('tab-listings'), text: 'สินค้า/บริการ')],
          ),
          Expanded(child: TabBarView(children: [_PartnerList(), _AdminListings()])),
        ]),
      );
}

class _PartnerList extends StatelessWidget {
  const _PartnerList();

  Future<void> _patch(BuildContext context, Json p, Map<String, Object?> body, String ok, Future<void> Function() reload) async {
    final api = AppScope.apiOf(context);
    if (await act(context, () => api.portal('PATCH', '/admin/v1/partners/${p['id']}', body: body), ok: ok)) await reload();
  }

  @override
  Widget build(BuildContext context) => PortalList(
        path: '/admin/v1/partners',
        searchHint: 'ค้นหาชื่อธุรกิจ',
        emptyText: 'ไม่พบพาร์ทเนอร์',
        filters: const [
          ListFilter('ทั้งหมด'),
          ListFilter('รออนุมัติ', {'status': 'pending'}),
          ListFilter('อนุมัติแล้ว', {'status': 'approved'}),
          ListFilter('ระงับ', {'status': 'suspended'}),
          ListFilter('ไม่อนุมัติ', {'status': 'rejected'}),
        ],
        itemBuilder: (context, p, reload) {
          final st = str(p['status']);
          return TepCard(
            key: Key('partner-${p['id']}'),
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Row(children: [
                Expanded(child: Text(str(p['businessName']), style: const TextStyle(fontWeight: FontWeight.w800))),
                StatePill(st, partnerStatusLabels),
              ]),
              Text('${str(p['ownerName'])} · ${str(p['ownerEmail'])}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
              if (str(p['description']).isNotEmpty) Padding(padding: const EdgeInsets.only(top: 4), child: Text(str(p['description']))),
              const SizedBox(height: 4),
              Text('คอมมิชชัน ${numOf(p['commissionRate'])}% · ${numOf(p['listingCount'])} รายการ · สมัคร ${dateOf(p['createdAt'])}', style: const TextStyle(fontSize: 12)),
              if (str(p['decisionNote']).isNotEmpty) Text('หมายเหตุ: ${str(p['decisionNote'])}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
              const SizedBox(height: 8),
              Wrap(spacing: 8, runSpacing: 8, children: [
                if (st == 'pending' || st == 'rejected' || st == 'suspended')
                  tileButton(st == 'suspended' ? 'เปิดใช้งาน' : 'อนุมัติ', key: Key('approve-${p['id']}'), () => _patch(context, p, {'status': 'approved'}, 'อนุมัติแล้ว', reload)),
                if (st == 'pending')
                  tileButton('ไม่อนุมัติ', key: Key('reject-${p['id']}'), danger: true, outlined: true, () async {
                    final note = await askText(context, title: 'ไม่อนุมัติ', label: 'เหตุผล (ผู้สมัครจะเห็นข้อความนี้)', maxLength: 500);
                    if (note == null || !context.mounted) return;
                    await _patch(context, p, {'status': 'rejected', 'decisionNote': note}, 'บันทึกแล้ว', reload);
                  }),
                if (st == 'approved')
                  tileButton('ระงับ', key: Key('suspend-${p['id']}'), danger: true, outlined: true, () async {
                    final note = await askText(context, title: 'ระงับพาร์ทเนอร์', label: 'เหตุผล (รายการของพาร์ทเนอร์จะถูกซ่อนจากแคตตาล็อก)', maxLength: 500);
                    if (note == null || !context.mounted) return;
                    await _patch(context, p, {'status': 'suspended', 'decisionNote': note}, 'ระงับแล้ว', reload);
                  }),
                tileButton('ปรับคอมมิชชัน', key: Key('commission-${p['id']}'), outlined: true, () async {
                  final v = await askText(context, title: 'คอมมิชชัน (%)', label: '0 – 50', initial: '${numOf(p['commissionRate'])}', maxLength: 5, keyboard: const TextInputType.numberWithOptions(decimal: true));
                  final rate = v == null ? null : double.tryParse(v);
                  if (rate == null || !context.mounted) return;
                  await _patch(context, p, {'commissionRate': rate}, 'บันทึกแล้ว', reload);
                }),
              ]),
            ]),
          );
        },
      );
}

class _AdminListings extends StatelessWidget {
  const _AdminListings();
  @override
  Widget build(BuildContext context) => PortalList(
        path: '/admin/v1/listings',
        searchHint: 'ค้นหาชื่อหรือสถานที่',
        emptyText: 'ไม่พบรายการ',
        filters: const [
          ListFilter('ทั้งหมด'),
          ListFilter('เปิดขาย', {'status': 'published'}),
          ListFilter('ฉบับร่าง', {'status': 'draft'}),
          ListFilter('เก็บถาวร', {'status': 'archived'}),
        ],
        itemBuilder: (context, l, reload) => TepCard(
          child: Row(children: [
            Icon(typeIcon(str(l['type'])), color: TepColors.primary),
            const SizedBox(width: 12),
            Expanded(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text(str(l['name']), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                Text('${str(l['location'])} · ${numOf(l['partnerId']) > 0 ? 'พาร์ทเนอร์ #${numOf(l['partnerId']).toInt()}' : 'แพลตฟอร์ม'}', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 12)),
              ]),
            ),
            StatePill(str(l['status']), listingStatusLabels),
            PopupMenuButton<String>(
              key: Key('listing-menu-${l['id']}'),
              tooltip: 'เปลี่ยนสถานะ',
              onSelected: (s) async {
                final api = AppScope.apiOf(context);
                if (await act(context, () => api.portal('PATCH', '/admin/v1/listings/${l['id']}', body: {'status': s}), ok: 'อัปเดตสถานะแล้ว')) await reload();
              },
              itemBuilder: (_) => [for (final s in listingStatusLabels.entries) PopupMenuItem(value: s.key, child: Text(s.value))],
            ),
          ]),
        ),
      );
}

/// Board 3.4 — system status and audit log.
class AdminSystemPage extends StatelessWidget {
  const AdminSystemPage({super.key});
  @override
  Widget build(BuildContext context) => DefaultTabController(
        length: 2,
        child: Column(children: [
          const TabBar(
            labelColor: TepColors.primary,
            indicatorColor: TepColors.primary,
            tabs: [Tab(key: Key('tab-system'), text: 'สถานะระบบ'), Tab(key: Key('tab-audit'), text: 'Audit log')],
          ),
          Expanded(
            child: TabBarView(children: [
              PortalPage(
                path: '/admin/v1/system',
                builder: (context, s) => [
                  TepCard(
                    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                      Text('เวอร์ชัน ${str(s['version'])} · ${str(s['env'])} · โหมด ${str(s['mode'])}', style: const TextStyle(fontWeight: FontWeight.w700)),
                      const SizedBox(height: 4),
                      Text('เวลาเซิร์ฟเวอร์ ${str(s['serverTime'])}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                      const Divider(height: 24),
                      for (final e in jmap(s['dependencies']).entries)
                        ListTile(
                          dense: true,
                          contentPadding: EdgeInsets.zero,
                          leading: Icon(e.value == 'ok' ? Icons.check_circle_rounded : (e.value == 'down' ? Icons.error_rounded : Icons.remove_circle_outline_rounded),
                              color: e.value == 'ok' ? TepColors.success : (e.value == 'down' ? TepColors.danger : TepColors.muted)),
                          title: Text(e.key),
                          trailing: Text(str(e.value)),
                        ),
                    ]),
                  ),
                ],
              ),
              PortalList(
                path: '/admin/v1/audit-logs',
                emptyText: 'ยังไม่มีบันทึก',
                filters: const [
                  ListFilter('ทั้งหมด'),
                  ListFilter('admin', {'action': 'admin.'}),
                  ListFilter('partner', {'action': 'partner.'}),
                  ListFilter('staff', {'action': 'staff.'}),
                ],
                itemBuilder: (context, a, reload) => TepCard(
                  child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text(str(a['action']), style: const TextStyle(fontWeight: FontWeight.w700)),
                    Text('${str(a['actorName'])} → ${str(a['entity'])} #${str(a['entityId'])}', style: const TextStyle(fontSize: 13)),
                    if (jmap(a['meta']).isNotEmpty) Text(jmap(a['meta']).toString(), style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                    Text(str(a['createdAt']), style: const TextStyle(color: TepColors.muted, fontSize: 11)),
                  ]),
                ),
              ),
            ]),
          ),
        ]),
      );
}
