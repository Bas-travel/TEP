import 'package:flutter/material.dart';

import '../../app_scope.dart';
import '../../core/format.dart';
import '../../core/theme.dart';
import '../../data/api_client.dart';
import '../widgets/common.dart';
import 'portal_kit.dart';

/// Board 4.1 — my work today.
class StaffDashboardPage extends StatelessWidget {
  const StaffDashboardPage({super.key});
  @override
  Widget build(BuildContext context) => PortalPage(
        path: '/staff/v1/dashboard',
        builder: (context, d) => [
          KpiGrid([
            Kpi('ตั๋วรอตอบ', '${numOf(d['openTickets'])}', Icons.support_agent_rounded),
            Kpi('การจองรอตรวจ', '${numOf(d['uncheckedBookings'])}', Icons.assignment_late_rounded),
            Kpi('ตรวจแล้ววันนี้', '${numOf(d['checkedToday'])}', Icons.verified_rounded),
            Kpi('พบความผิดปกติวันนี้', '${numOf(d['flaggedToday'])}', Icons.flag_rounded),
            Kpi('ตอบครั้งแรกเฉลี่ย', '${numOf(d['avgFirstResponseHours'])} ชม.', Icons.timer_rounded),
          ]),
        ],
      );
}

/// Board 4.2 — verify bookings.
class StaffBookingsPage extends StatelessWidget {
  const StaffBookingsPage({super.key});
  @override
  Widget build(BuildContext context) => PortalList(
        path: '/staff/v1/bookings',
        searchHint: 'ค้นหาเลขที่การจองหรือชื่อ',
        emptyText: 'ไม่พบการจอง',
        filters: const [
          ListFilter('รอตรวจ', {'status': 'confirmed', 'checked': 'false'}),
          ListFilter('ตรวจแล้ว', {'checked': 'true'}),
          ListFilter('ทั้งหมด'),
        ],
        itemBuilder: (context, b, reload) => TepCard(
          key: Key('sbooking-${b['id']}'),
          onTap: () async {
            final changed = await showModalBottomSheet<bool>(context: context, isScrollControlled: true, builder: (_) => _BookingSheet(booking: b));
            if (changed == true) await reload();
          },
          child: Row(children: [
            Icon(typeIcon(str(b['type'])), color: TepColors.primary),
            const SizedBox(width: 12),
            Expanded(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text(str(b['title']), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                Text('${str(b['bookingNumber'])} · ${str(b['travelerName'])}', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                Text(thb(numOf(b['amount'])), style: const TextStyle(fontSize: 13)),
              ]),
            ),
            Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
              StatusBadge(str(b['status'])),
              const SizedBox(height: 4),
              StatePill(str(b['verification']), verificationLabels),
            ]),
          ]),
        ),
      );
}

class _BookingSheet extends StatelessWidget {
  const _BookingSheet({required this.booking});
  final Json booking;

  Future<void> _check(BuildContext context, String result) async {
    final api = AppScope.apiOf(context);
    String note = '';
    if (result == 'flagged') {
      final n = await askText(context, title: 'ทำเครื่องหมายผิดปกติ', label: 'รายละเอียดที่พบ', maxLength: 500);
      if (n == null || !context.mounted) return;
      note = n;
    }
    final ok = await act(context, () => api.portal('POST', '/staff/v1/bookings/${booking['id']}/check', body: {'result': result, if (note.isNotEmpty) 'note': note}),
        ok: result == 'verified' ? 'บันทึกการตรวจสอบแล้ว' : 'บันทึกความผิดปกติแล้ว');
    if (ok && context.mounted) Navigator.of(context).pop(true);
  }

  @override
  Widget build(BuildContext context) {
    final b = booking;
    final checks = jlist(b['checks']);
    final paid = b['status'] == 'confirmed' || b['status'] == 'completed';
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(str(b['title']), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800)),
          Text('${str(b['bookingNumber'])} · ${dateOf(b['date'])}', style: const TextStyle(color: TepColors.muted)),
          const SizedBox(height: 8),
          Text('ผู้เดินทาง: ${str(b['travelerName'])} (${str(b['travelerEmail'])})'),
          Text('ยอดชำระ ${thb(numOf(b['amount']))} · วิธีชำระ ${str(b['method']).isEmpty ? '-' : str(b['method'])} · ${numOf(b['guests'])} ท่าน'),
          const SizedBox(height: 12),
          const Text('ประวัติการตรวจสอบ', style: TextStyle(fontWeight: FontWeight.w700)),
          if (checks.isEmpty) const Text('ยังไม่เคยตรวจสอบ', style: TextStyle(color: TepColors.muted)),
          for (final c in checks)
            ListTile(
              dense: true,
              contentPadding: EdgeInsets.zero,
              leading: Icon(c['result'] == 'verified' ? Icons.verified_rounded : Icons.flag_rounded, color: toneOf(str(c['result']))),
              title: Text('${verificationLabels[c['result']] ?? str(c['result'])} · ${str(c['staffName'])}'),
              subtitle: Text('${str(c['note'])} ${dateOf(c['createdAt'])}'.trim()),
            ),
          const SizedBox(height: 12),
          if (paid)
            Row(children: [
              Expanded(child: tileButton('ถูกต้อง', key: const Key('check-verified'), () => _check(context, 'verified'))),
              const SizedBox(width: 8),
              Expanded(child: tileButton('ผิดปกติ', key: const Key('check-flagged'), danger: true, outlined: true, () => _check(context, 'flagged'))),
            ])
          else
            const Text('ตรวจสอบได้เฉพาะการจองที่ชำระเงินแล้ว', style: TextStyle(color: TepColors.muted, fontSize: 12)),
        ]),
      ),
    );
  }
}

/// Board 4.3 — support tickets (replaces live chat in the MVP).
class StaffTicketsPage extends StatelessWidget {
  const StaffTicketsPage({super.key});
  @override
  Widget build(BuildContext context) => PortalList(
        path: '/staff/v1/tickets',
        searchHint: 'ค้นหาหัวข้อ ข้อความ หรืออีเมล',
        emptyText: 'ไม่มีตั๋ว',
        filters: const [
          ListFilter('ทั้งหมด'),
          ListFilter('รอตอบ', {'status': 'open'}),
          ListFilter('ตอบแล้ว', {'status': 'answered'}),
          ListFilter('ปิดเรื่อง', {'status': 'resolved'}),
        ],
        itemBuilder: (context, t, reload) => TepCard(
          key: Key('ticket-${t['id']}'),
          onTap: () async {
            await Navigator.of(context).push<void>(MaterialPageRoute<void>(builder: (_) => TicketDetailScreen(ticketId: (t['id'] as num).toInt())));
            await reload();
          },
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Expanded(child: Text(str(t['subject']), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700))),
              StatePill(str(t['status']), ticketStatusLabels),
            ]),
            Text(str(t['message']), maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted)),
            Text('${str(t['travelerName'])} · ${dateOf(t['createdAt'])} · ตอบแล้ว ${numOf(t['replyCount'])} ครั้ง', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
          ]),
        ),
      );
}

class TicketDetailScreen extends StatefulWidget {
  const TicketDetailScreen({super.key, required this.ticketId});
  final int ticketId;
  @override
  State<TicketDetailScreen> createState() => _TicketDetailScreenState();
}

class _TicketDetailScreenState extends State<TicketDetailScreen> {
  final _reply = TextEditingController();
  int _version = 0;
  bool _busy = false;

  @override
  void dispose() {
    _reply.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final api = AppScope.apiOf(context);
    setState(() => _busy = true);
    final ok = await act(context, () => api.portal('POST', '/staff/v1/tickets/${widget.ticketId}/replies', body: {'message': _reply.text.trim()}), ok: 'ส่งคำตอบแล้ว');
    if (!mounted) return;
    setState(() {
      _busy = false;
      if (ok) {
        _reply.clear();
        _version++;
      }
    });
  }

  Future<void> _setStatus(String status) async {
    final api = AppScope.apiOf(context);
    if (await act(context, () => api.portal('PATCH', '/staff/v1/tickets/${widget.ticketId}', body: {'status': status}), ok: status == 'resolved' ? 'ปิดเรื่องแล้ว' : 'เปิดเรื่องใหม่แล้ว') && mounted) {
      setState(() => _version++);
    }
  }

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: Text('ตั๋ว #${widget.ticketId}')),
      body: AsyncBody<Json>(
        key: ValueKey('tk-$_version'),
        load: () => api.portal('GET', '/staff/v1/tickets/${widget.ticketId}'),
        builder: (context, t, reload) {
          final replies = jlist(t['replies']);
          final resolved = t['status'] == 'resolved';
          return Column(children: [
            Expanded(
              child: ListView(padding: const EdgeInsets.all(16), children: [
                Row(children: [
                  Expanded(child: Text(str(t['subject']), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800))),
                  StatePill(str(t['status']), ticketStatusLabels),
                ]),
                Text('${str(t['travelerName'])} · ${str(t['travelerEmail'])}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                const SizedBox(height: 12),
                _Bubble(text: str(t['message']), who: 'ผู้เดินทาง', mine: false, at: str(t['createdAt'])),
                for (final r in replies) _Bubble(text: str(r['message']), who: '${str(r['authorName'])} (${r['authorRole'] == 'traveler' ? 'ผู้เดินทาง' : 'ทีมงาน'})', mine: r['authorRole'] != 'traveler', at: str(r['createdAt'])),
              ]),
            ),
            SafeArea(
              top: false,
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
                child: Column(mainAxisSize: MainAxisSize.min, children: [
                  if (!resolved)
                    Row(children: [
                      Expanded(child: TextField(key: const Key('reply-text'), controller: _reply, maxLines: 3, minLines: 1, maxLength: 2000, decoration: const InputDecoration(hintText: 'พิมพ์คำตอบ…', counterText: ''))),
                      IconButton(key: const Key('reply-send'), onPressed: _busy ? null : _send, icon: const Icon(Icons.send_rounded, color: TepColors.primary)),
                    ]),
                  const SizedBox(height: 8),
                  tileButton(resolved ? 'เปิดเรื่องอีกครั้ง' : 'ปิดเรื่อง (แก้ไขแล้ว)', key: Key(resolved ? 'ticket-reopen' : 'ticket-resolve'), outlined: true, () => _setStatus(resolved ? 'open' : 'resolved')),
                ]),
              ),
            ),
          ]);
        },
      ),
    );
  }
}

class _Bubble extends StatelessWidget {
  const _Bubble({required this.text, required this.who, required this.mine, required this.at});
  final String text, who, at;
  final bool mine;
  @override
  Widget build(BuildContext context) => Align(
        alignment: mine ? Alignment.centerRight : Alignment.centerLeft,
        child: Container(
          margin: const EdgeInsets.only(bottom: 8),
          padding: const EdgeInsets.all(12),
          constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.8),
          decoration: BoxDecoration(color: mine ? TepColors.secondary.withValues(alpha: 0.3) : Colors.white, borderRadius: BorderRadius.circular(14), border: Border.all(color: TepColors.line)),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(who, style: const TextStyle(color: TepColors.muted, fontSize: 11)),
            Text(text),
            Text(dateOf(at), style: const TextStyle(color: TepColors.muted, fontSize: 11)),
          ]),
        ),
      );
}

/// Board 4.4 — daily report (Asia/Bangkok day).
class StaffReportPage extends StatefulWidget {
  const StaffReportPage({super.key});
  @override
  State<StaffReportPage> createState() => _StaffReportPageState();
}

class _StaffReportPageState extends State<StaffReportPage> {
  DateTime? _date;

  Future<void> _pick() async {
    final now = DateTime.now();
    final d = await showDatePicker(context: context, initialDate: _date ?? now, firstDate: DateTime(now.year - 1), lastDate: now);
    if (d != null && mounted) setState(() => _date = d);
  }

  @override
  Widget build(BuildContext context) => PortalPage(
        key: ValueKey('report-$_date'),
        path: '/staff/v1/reports/daily',
        query: {'date': _date == null ? null : isoDate(_date!)},
        builder: (context, r) => [
          Row(children: [
            Expanded(child: Text('รายงานประจำวัน ${str(r['date'])}', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800))),
            tileButton('เลือกวัน', key: const Key('report-pick'), outlined: true, _pick),
          ]),
          const SizedBox(height: 12),
          KpiGrid([
            Kpi('การจองใหม่', '${numOf(r['bookingsCreated'])}', Icons.receipt_long_rounded),
            Kpi('ตรวจแล้ว (ถูกต้อง)', '${numOf(r['bookingsVerified'])}', Icons.verified_rounded),
            Kpi('ผิดปกติ', '${numOf(r['bookingsFlagged'])}', Icons.flag_rounded),
            Kpi('ที่ฉันตรวจ', '${numOf(r['myChecks'])}', Icons.person_rounded),
            Kpi('ตั๋วใหม่', '${numOf(r['ticketsOpened'])}', Icons.mark_email_unread_rounded),
            Kpi('ตอบครั้งแรก', '${numOf(r['ticketsFirstAnswered'])}', Icons.reply_rounded),
            Kpi('ปิดเรื่อง', '${numOf(r['ticketsResolved'])}', Icons.task_alt_rounded),
          ]),
          const SizedBox(height: 8),
          Text('เขตเวลา ${str(r['timezone'])}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
        ],
      );
}
