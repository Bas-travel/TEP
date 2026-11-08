import 'package:flutter/material.dart';

import '../../app_scope.dart';
import '../../core/format.dart';
import '../../core/theme.dart';
import '../../data/api_client.dart';
import '../../data/models.dart';
import '../widgets/common.dart';
import 'portal_kit.dart';

/// Board 2.1 — partner dashboard.
class PartnerDashboardPage extends StatelessWidget {
  const PartnerDashboardPage({super.key});
  @override
  Widget build(BuildContext context) => PortalPage(
        path: '/partner/v1/dashboard',
        builder: (context, d) {
          final by = jmap(d['listings']);
          final month = jmap(d['thisMonth']);
          return [
            Text(str(d['businessName']), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800)),
            const SizedBox(height: 12),
            KpiGrid([
              Kpi('เปิดขายอยู่', '${numOf(by['published'])}', Icons.storefront_rounded),
              Kpi('รอให้เสร็จสิ้น', '${numOf(d['bookingsToComplete'])}', Icons.hourglass_bottom_rounded),
              Kpi('รอชำระเงิน', '${numOf(d['bookingsAwaitingPayment'])}', Icons.schedule_rounded),
              Kpi('รายได้สุทธิเดือนนี้', thb(numOf(month['net'])), Icons.payments_rounded),
              Kpi('ฉบับร่าง / เก็บถาวร', '${numOf(by['draft'])} / ${numOf(by['archived'])}', Icons.inventory_2_rounded),
              Kpi('คอมมิชชัน', '${numOf(d['commissionRate'])}%', Icons.percent_rounded),
            ]),
          ];
        },
      );
}

/// Board 2.2 — manage products / services.
class PartnerListingsPage extends StatefulWidget {
  const PartnerListingsPage({super.key});
  @override
  State<PartnerListingsPage> createState() => _PartnerListingsPageState();
}

class _PartnerListingsPageState extends State<PartnerListingsPage> {
  int _version = 0;

  Future<void> _open(Json? item) async {
    final saved = await Navigator.of(context).push<bool>(MaterialPageRoute<bool>(builder: (_) => ListingFormScreen(existing: item)));
    if (saved == true && mounted) setState(() => _version++);
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        backgroundColor: Colors.transparent,
        floatingActionButton: FloatingActionButton.extended(
          key: const Key('listing-add'),
          onPressed: () => _open(null),
          icon: const Icon(Icons.add_rounded),
          label: const Text('เพิ่มรายการ'),
        ),
        body: PortalList(
          key: ValueKey('listings-$_version'),
          path: '/partner/v1/listings',
          searchHint: 'ค้นหาชื่อหรือสถานที่',
          emptyText: 'ยังไม่มีรายการ — กด "เพิ่มรายการ" เพื่อเริ่มต้น',
          filters: const [
            ListFilter('ทั้งหมด'),
            ListFilter('เปิดขาย', {'status': 'published'}),
            ListFilter('ฉบับร่าง', {'status': 'draft'}),
            ListFilter('เก็บถาวร', {'status': 'archived'}),
          ],
          itemBuilder: (context, l, reload) => TepCard(
            onTap: () => _open(l),
            child: Row(children: [
              Icon(typeIcon(str(l['type'])), color: TepColors.primary),
              const SizedBox(width: 12),
              Expanded(
                child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                  Text(str(l['name']), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                  Text('${str(l['location'])} · เริ่ม ${thb(numOf(l['priceFrom']))}', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 13)),
                ]),
              ),
              StatePill(str(l['status']), listingStatusLabels),
            ]),
          ),
        ),
      );
}

class ListingFormScreen extends StatefulWidget {
  const ListingFormScreen({super.key, this.existing});
  final Json? existing;
  @override
  State<ListingFormScreen> createState() => _ListingFormScreenState();
}

class _ListingFormScreenState extends State<ListingFormScreen> {
  late final Json _e = widget.existing ?? <String, Object?>{};
  late final _name = TextEditingController(text: str(_e['name']));
  late final _location = TextEditingController(text: str(_e['location']));
  late final _price = TextEditingController(text: _e['priceFrom'] == null ? '' : '${_e['priceFrom']}');
  late final _desc = TextEditingController(text: str(_e['description']));
  late final _amen = TextEditingController(text: ((_e['amenities'] as List?) ?? const []).join(', '));
  late final _image = TextEditingController(text: str(_e['imageUrl']));
  late String _type = _e['type'] is String && const ['hotel', 'restaurant', 'attraction'].contains(_e['type']) ? _e['type'] as String : 'hotel';
  late String _status = _e['status'] is String ? _e['status'] as String : 'draft';
  int? _dest;
  List<Destination> _dests = const [];
  bool _busy = false;

  @override
  void initState() {
    super.initState();
    _dest = _e['destinationId'] is num ? (_e['destinationId'] as num).toInt() : null;
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      final api = AppScope.apiOf(context);
      final ds = await guarded(context, api.destinations);
      if (!mounted || ds == null) return;
      setState(() {
        _dests = ds;
        _dest ??= ds.isEmpty ? null : ds.first.id;
      });
    });
  }

  @override
  void dispose() {
    for (final c in [_name, _location, _price, _desc, _amen, _image]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    final api = AppScope.apiOf(context);
    final price = double.tryParse(_price.text.trim());
    final body = <String, Object?>{
      'name': _name.text.trim(),
      'type': _type,
      'location': _location.text.trim(),
      'destinationId': _dest ?? 0,
      'priceFrom': price ?? 0,
      'description': _desc.text.trim(),
      'amenities': [for (final a in _amen.text.split(',')) if (a.trim().isNotEmpty) a.trim()],
      'imageUrl': _image.text.trim(),
      'status': _status,
    };
    setState(() => _busy = true);
    final id = _e['id'];
    final ok = await act(
      context,
      () => id == null ? api.portal('POST', '/partner/v1/listings', body: body) : api.portal('PATCH', '/partner/v1/listings/$id', body: body),
      ok: 'บันทึกแล้ว',
    );
    if (!mounted) return;
    setState(() => _busy = false);
    if (ok) Navigator.of(context).pop(true);
  }

  @override
  Widget build(BuildContext context) {
    final isNew = _e['id'] == null;
    return Scaffold(
      appBar: AppBar(title: Text(isNew ? 'เพิ่มรายการ' : 'แก้ไขรายการ')),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        TextField(key: const Key('f-name'), controller: _name, maxLength: 190, decoration: const InputDecoration(labelText: 'ชื่อ')),
        const SizedBox(height: 8),
        DropdownButtonFormField<String>(
          key: const Key('f-type'),
          initialValue: _type,
          decoration: const InputDecoration(labelText: 'ประเภท'),
          items: [for (final t in const ['hotel', 'restaurant', 'attraction']) DropdownMenuItem(value: t, child: Text(typeLabels[t] ?? t))],
          onChanged: (v) => setState(() => _type = v ?? _type),
        ),
        const SizedBox(height: 12),
        TextField(key: const Key('f-location'), controller: _location, maxLength: 190, decoration: const InputDecoration(labelText: 'สถานที่')),
        const SizedBox(height: 8),
        DropdownButtonFormField<int>(
          key: ValueKey('f-dest-${_dests.length}'),
          initialValue: _dests.any((d) => d.id == _dest) ? _dest : null,
          decoration: const InputDecoration(labelText: 'จุดหมายปลายทาง'),
          items: [for (final d in _dests) DropdownMenuItem(value: d.id, child: Text(d.name))],
          onChanged: (v) => setState(() => _dest = v),
        ),
        const SizedBox(height: 12),
        TextField(key: const Key('f-price'), controller: _price, keyboardType: const TextInputType.numberWithOptions(decimal: true), decoration: const InputDecoration(labelText: 'ราคาเริ่มต้น (บาท)')),
        const SizedBox(height: 12),
        TextField(key: const Key('f-desc'), controller: _desc, maxLines: 4, maxLength: 2000, decoration: const InputDecoration(labelText: 'รายละเอียด')),
        const SizedBox(height: 8),
        TextField(key: const Key('f-amen'), controller: _amen, decoration: const InputDecoration(labelText: 'สิ่งอำนวยความสะดวก (คั่นด้วย ,)')),
        const SizedBox(height: 12),
        TextField(key: const Key('f-image'), controller: _image, keyboardType: TextInputType.url, decoration: const InputDecoration(labelText: 'ลิงก์รูปภาพ (https://…)')),
        const SizedBox(height: 12),
        DropdownButtonFormField<String>(
          key: const Key('f-status'),
          initialValue: _status,
          decoration: const InputDecoration(labelText: 'สถานะ'),
          items: [for (final s in listingStatusLabels.entries) DropdownMenuItem(value: s.key, child: Text(s.value))],
          onChanged: (v) => setState(() => _status = v ?? _status),
        ),
        const SizedBox(height: 4),
        const Text('เฉพาะสถานะ "เปิดขาย" ที่นักท่องเที่ยวมองเห็นและจองได้', style: TextStyle(color: TepColors.muted, fontSize: 12)),
        const SizedBox(height: 20),
        FilledButton(key: const Key('f-save'), onPressed: _busy ? null : _save, child: const Text('บันทึก')),
      ]),
    );
  }
}

/// Board 2.3 — bookings. The board's "ยืนยัน/ปฏิเสธ" maps to complete / reject (docs/requirements/mvp-scope.md).
class PartnerBookingsPage extends StatelessWidget {
  const PartnerBookingsPage({super.key});
  @override
  Widget build(BuildContext context) => PortalList(
        path: '/partner/v1/bookings',
        searchHint: 'ค้นหาเลขที่การจองหรือชื่อ',
        emptyText: 'ยังไม่มีการจอง',
        filters: const [
          ListFilter('ทั้งหมด'),
          ListFilter('รอชำระ', {'status': 'pending_payment'}),
          ListFilter('ยืนยันแล้ว', {'status': 'confirmed'}),
          ListFilter('เสร็จสิ้น', {'status': 'completed'}),
          ListFilter('ยกเลิก', {'status': 'cancelled'}),
        ],
        itemBuilder: (context, b, reload) {
          final api = AppScope.apiOf(context);
          final id = b['id'];
          final status = str(b['status']);
          return TepCard(
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Row(children: [
                Expanded(child: Text(str(b['title']), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700))),
                StatusBadge(status),
              ]),
              Text('${str(b['bookingNumber'])} · ${str(b['travelerName'])} · ${dateOf(b['date'])}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
              const SizedBox(height: 6),
              Text('ยอดขาย ${thb(numOf(b['gross']))}  ·  คอมมิชชัน ${thb(numOf(b['commission']))}  ·  สุทธิ ${thb(numOf(b['net']))}', style: const TextStyle(fontSize: 13)),
              if (status == 'cancelled' && str(b['cancelReason']).isNotEmpty) Text(str(b['cancelReason']), style: const TextStyle(color: TepColors.danger, fontSize: 12)),
              if (b['canAct'] == true) ...[
                const SizedBox(height: 10),
                Wrap(spacing: 8, children: [
                  if (status == 'confirmed')
                    tileButton('เสร็จสิ้นการให้บริการ', key: Key('complete-$id'), () async {
                      if (!await confirmDialog(context, 'เสร็จสิ้นการให้บริการ', 'ยืนยันว่าให้บริการตามการจอง ${str(b['bookingNumber'])} แล้ว?')) return;
                      if (!context.mounted) return;
                      if (await act(context, () => api.portal('POST', '/partner/v1/bookings/$id/complete'), ok: 'อัปเดตสถานะแล้ว')) await reload();
                    }),
                  tileButton('ปฏิเสธ', key: Key('reject-$id'), danger: true, outlined: true, () async {
                    final reason = await askText(context, title: 'ปฏิเสธการจอง', label: 'เหตุผล (ผู้เดินทางจะได้รับเงินคืนถ้าชำระแล้ว)');
                    if (reason == null || !context.mounted) return;
                    if (await act(context, () => api.portal('POST', '/partner/v1/bookings/$id/reject', body: {'reason': reason}), ok: 'ปฏิเสธและคืนเงินแล้ว')) await reload();
                  }),
                ]),
              ] else if (status == 'confirmed' || status == 'pending_payment')
                const Padding(padding: EdgeInsets.only(top: 6), child: Text('การจองนี้มีรายการของผู้ให้บริการอื่น — ติดต่อฝ่ายช่วยเหลือเพื่อเปลี่ยนสถานะ', style: TextStyle(color: TepColors.muted, fontSize: 12))),
            ]),
          );
        },
      );
}

/// Board 2.4 — earnings & commission, plus the business profile.
class PartnerEarningsPage extends StatefulWidget {
  const PartnerEarningsPage({super.key});
  @override
  State<PartnerEarningsPage> createState() => _PartnerEarningsPageState();
}

class _PartnerEarningsPageState extends State<PartnerEarningsPage> {
  int _version = 0;

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return AsyncBody<List<Json>>(
      key: ValueKey('earn-$_version'),
      load: () => Future.wait([api.portal('GET', '/partner/v1/earnings'), api.portal('GET', '/partner/v1/profile')]),
      builder: (context, data, reload) {
        final e = data[0], p = data[1];
        final months = jlist(e['months']);
        return RefreshIndicator(
          onRefresh: reload,
          child: ListView(physics: const AlwaysScrollableScrollPhysics(), padding: const EdgeInsets.all(16), children: [
            TepCard(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Row(children: [
                  Expanded(child: Text(str(p['businessName']), style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16))),
                  StatePill(str(p['status']), partnerStatusLabels),
                ]),
                if (str(p['contactPhone']).isNotEmpty) Text('โทร ${str(p['contactPhone'])}', style: const TextStyle(color: TepColors.muted)),
                if (str(p['description']).isNotEmpty) Text(str(p['description'])),
                const SizedBox(height: 8),
                tileButton('แก้ไขโปรไฟล์ธุรกิจ', key: const Key('profile-edit'), outlined: true, () async {
                  final saved = await Navigator.of(context).push<bool>(MaterialPageRoute<bool>(builder: (_) => PartnerProfileForm(profile: p)));
                  if (saved == true && mounted) setState(() => _version++);
                }),
              ]),
            ),
            const SizedBox(height: 12),
            KpiGrid([
              Kpi('ยอดขายรวม', thb(numOf(e['gross'])), Icons.trending_up_rounded),
              Kpi('คอมมิชชัน ${numOf(e['commissionRate'])}%', thb(numOf(e['commission'])), Icons.percent_rounded),
              Kpi('รายได้สุทธิ', thb(numOf(e['net'])), Icons.account_balance_wallet_rounded),
              Kpi('เดือนที่มีรายการ', '${months.length}', Icons.calendar_month_rounded),
            ]),
            const SectionHeader('รายเดือน'),
            if (months.isEmpty) const EmptyState(Icons.receipt_long_outlined, 'ยังไม่มีรายได้'),
            for (final m in months)
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: TepCard(
                  child: Row(children: [
                    Expanded(child: Text('${str(m['month'])} · ${numOf(m['bookings'])} การจอง', style: const TextStyle(fontWeight: FontWeight.w700))),
                    Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
                      Text('สุทธิ ${thb(numOf(m['net']))}', style: const TextStyle(color: TepColors.success, fontWeight: FontWeight.w700)),
                      Text('ขาย ${thb(numOf(m['gross']))} · หัก ${thb(numOf(m['commission']))}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                    ]),
                  ]),
                ),
              ),
            const Padding(
              padding: EdgeInsets.only(top: 8),
              child: Text('คำนวณจากการจองที่ยืนยันแล้วและเสร็จสิ้น การโอนเงินรอบจ่าย (payout) ยังไม่อยู่ในขอบเขต MVP', style: TextStyle(color: TepColors.muted, fontSize: 12)),
            ),
          ]),
        );
      },
    );
  }
}

class PartnerProfileForm extends StatefulWidget {
  const PartnerProfileForm({super.key, required this.profile});
  final Json profile;
  @override
  State<PartnerProfileForm> createState() => _PartnerProfileFormState();
}

class _PartnerProfileFormState extends State<PartnerProfileForm> {
  late final _name = TextEditingController(text: str(widget.profile['businessName']));
  late final _phone = TextEditingController(text: str(widget.profile['contactPhone']));
  late final _desc = TextEditingController(text: str(widget.profile['description']));
  bool _busy = false;

  @override
  void dispose() {
    _name.dispose();
    _phone.dispose();
    _desc.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final api = AppScope.apiOf(context);
    setState(() => _busy = true);
    final ok = await act(
      context,
      () => api.portal('PATCH', '/partner/v1/profile', body: {'businessName': _name.text.trim(), 'contactPhone': _phone.text.trim(), 'description': _desc.text.trim()}),
      ok: 'บันทึกโปรไฟล์แล้ว',
    );
    if (!mounted) return;
    setState(() => _busy = false);
    if (ok) Navigator.of(context).pop(true);
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('โปรไฟล์ธุรกิจ')),
        body: ListView(padding: const EdgeInsets.all(16), children: [
          TextField(key: const Key('pp-name'), controller: _name, maxLength: 120, decoration: const InputDecoration(labelText: 'ชื่อธุรกิจ')),
          const SizedBox(height: 8),
          TextField(key: const Key('pp-phone'), controller: _phone, keyboardType: TextInputType.phone, maxLength: 30, decoration: const InputDecoration(labelText: 'เบอร์ติดต่อ')),
          const SizedBox(height: 8),
          TextField(key: const Key('pp-desc'), controller: _desc, maxLines: 4, maxLength: 1000, decoration: const InputDecoration(labelText: 'รายละเอียดธุรกิจ')),
          const SizedBox(height: 16),
          FilledButton(key: const Key('pp-save'), onPressed: _busy ? null : _save, child: const Text('บันทึก')),
        ]),
      );
}
