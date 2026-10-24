import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'widgets/common.dart';

/// Board 1.10 — My bookings with status filter; 1.11 — detail + cancel.
class BookingsScreen extends StatefulWidget {
  const BookingsScreen({super.key});
  @override
  State<BookingsScreen> createState() => _BookingsScreenState();
}

class _BookingsScreenState extends State<BookingsScreen> {
  String _status = 'all';
  int _version = 0;

  static const _filters = [('all', 'ทั้งหมด'), ('pending_payment', 'รอชำระ'), ('confirmed', 'ยืนยันแล้ว'), ('cancelled', 'ยกเลิก')];

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('การจองของฉัน'), actions: [
        IconButton(tooltip: 'รีเฟรช', onPressed: () => setState(() => _version++), icon: const Icon(Icons.refresh_rounded)),
      ]),
      body: Column(children: [
        SizedBox(
          height: 52,
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            children: [
              for (final f in _filters)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 8),
                  child: ChoiceChip(key: Key('bf-${f.$1}'), label: Text(f.$2), selected: _status == f.$1, onSelected: (_) => setState(() => _status = f.$1)),
                ),
            ],
          ),
        ),
        Expanded(
          child: AsyncBody<List<BookingSummary>>(
            key: ValueKey('$_status-$_version'),
            load: () => api.bookings(status: _status == 'all' ? null : _status),
            builder: (context, items, reload) {
              if (items.isEmpty) return const EmptyState(Icons.receipt_long_outlined, 'ยังไม่มีการจอง');
              return RefreshIndicator(
                onRefresh: reload,
                child: ListView.separated(
                  padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
                  itemCount: items.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 10),
                  itemBuilder: (context, i) {
                    final b = items[i];
                    return TepCard(
                      onTap: () async {
                        await Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => BookingDetailScreen(bookingId: b.id)));
                        if (mounted) setState(() => _version++);
                      },
                      child: Row(children: [
                        ListingThumb(type: b.type, seed: b.id, size: 56),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                            Text(b.title, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                            Text('${b.bookingNumber} · ${dmy(b.date)}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                            const SizedBox(height: 6),
                            Row(children: [
                              StatusBadge(b.status),
                              const Spacer(),
                              Text(thb(b.amount), style: const TextStyle(fontWeight: FontWeight.w800, color: TepColors.navyBar)),
                            ]),
                          ]),
                        ),
                      ]),
                    );
                  },
                ),
              );
            },
          ),
        ),
      ]),
    );
  }
}

class BookingDetailScreen extends StatefulWidget {
  const BookingDetailScreen({super.key, required this.bookingId});
  final int bookingId;
  @override
  State<BookingDetailScreen> createState() => _BookingDetailScreenState();
}

class _BookingDetailScreenState extends State<BookingDetailScreen> {
  int _version = 0;

  Future<void> _cancel(BookingDetail b) async {
    final reason = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('ยกเลิกการจอง'),
        content: TextField(key: const Key('cancel-reason'), controller: reason, decoration: const InputDecoration(labelText: 'เหตุผล')),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('ไม่ยกเลิก')),
          TextButton(key: const Key('cancel-confirm'), onPressed: () => Navigator.pop(ctx, true), child: const Text('ยืนยันยกเลิก')),
        ],
      ),
    );
    final text = reason.text;
    reason.dispose();
    if (ok != true || !mounted) return;
    final api = AppScope.apiOf(context);
    final res = await guarded(context, () => api.cancelBooking(b.id, text));
    if (res != null && mounted) setState(() => _version++);
  }

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('รายละเอียดการจอง')),
      body: AsyncBody<BookingDetail>(
        key: ValueKey(_version),
        load: () => api.booking(widget.bookingId),
        builder: (context, b, _) => ListView(padding: const EdgeInsets.all(16), children: [
          TepCard(
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Row(children: [Expanded(child: Text(b.bookingNumber, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 18))), StatusBadge(b.status)]),
              const SizedBox(height: 8),
              Text(b.title, style: const TextStyle(fontWeight: FontWeight.w700)),
              Text(b.subtitle, style: const TextStyle(color: TepColors.muted)),
              const Divider(height: 28),
              _kv('วันที่จอง', dmy(b.date)),
              _kv('วิธีชำระเงิน', b.method.isEmpty ? '-' : b.method),
              _kv('ราคา', thb(b.subtotal)),
              _kv('ค่าบริการ', thb(b.serviceFee)),
              _kv('ยอดรวม', thb(b.amount), bold: true),
              if (b.status == 'cancelled') ...[
                const Divider(height: 28),
                _kv('เหตุผลที่ยกเลิก', b.cancelReason ?? '-'),
                _kv('เงินคืนเข้า Wallet', thb(b.refundAmount ?? 0)),
              ],
            ]),
          ),
          if (b.status == 'pending_payment' || b.status == 'confirmed') ...[
            const SizedBox(height: 16),
            OutlinedButton(
              key: const Key('cancel-booking'),
              style: OutlinedButton.styleFrom(minimumSize: const Size.fromHeight(50), foregroundColor: TepColors.danger),
              onPressed: () => _cancel(b),
              child: const Text('ยกเลิกการจอง'),
            ),
          ],
        ]),
      ),
    );
  }

  Widget _kv(String k, String v, {bool bold = false}) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 4),
        child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Expanded(child: Text(k, style: const TextStyle(color: TepColors.muted))),
          Flexible(child: Text(v, textAlign: TextAlign.end, style: TextStyle(fontWeight: bold ? FontWeight.w800 : FontWeight.w600))),
        ]),
      );
}
