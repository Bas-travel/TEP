import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/api_client.dart';
import '../data/models.dart';
import 'widgets/common.dart';

String newIdempotencyKey() => 'k${DateTime.now().microsecondsSinceEpoch}-${identityHashCode(Object())}';

/// Board 1.5–1.9 — choose date/guests, review price, pay (PromptPay / card / Wallet), success.
class BookingFlowScreen extends StatefulWidget {
  const BookingFlowScreen({super.key, required this.listing, this.pollInterval = const Duration(seconds: 1)});
  final ListingDetail listing;
  final Duration pollInterval;
  @override
  State<BookingFlowScreen> createState() => _BookingFlowScreenState();
}

enum _Step { details, payment, success }

class _BookingFlowScreenState extends State<BookingFlowScreen> {
  _Step _step = _Step.details;
  DateTime _date = DateTime.now().add(const Duration(days: 7));
  int _qty = 1;
  final _coupon = TextEditingController();
  String _method = 'promptpay';
  bool _busy = false;
  BookingDetail? _booking;
  PaymentIntent? _pending;
  late final String _bookingKey = newIdempotencyKey();
  String _intentKey = newIdempotencyKey();

  @override
  void dispose() {
    _coupon.dispose();
    super.dispose();
  }

  double get _subtotal => widget.listing.priceFrom * _qty;

  Future<void> _pickDate() async {
    final d = await showDatePicker(
      context: context,
      initialDate: _date,
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );
    if (d != null) setState(() => _date = d);
  }

  Future<void> _createBooking() async {
    final api = AppScope.apiOf(context);
    setState(() => _busy = true);
    final b = await guarded(
      context,
      () => api.createBooking(lines: [BookingLine(widget.listing.id, _qty)], couponCode: _coupon.text.trim(), idempotencyKey: _bookingKey),
    );
    if (!mounted) return;
    setState(() {
      _busy = false;
      if (b != null) {
        _booking = b;
        _step = _Step.payment;
      }
    });
  }

  Future<void> _pay() async {
    final api = AppScope.apiOf(context);
    final b = _booking!;
    setState(() => _busy = true);
    var pi = await guarded(context, () => api.createIntent(bookingId: b.id, method: _method, idempotencyKey: _intentKey));
    if (!mounted) return;
    if (pi == null) {
      setState(() => _busy = false);
      return;
    }
    setState(() => _pending = pi);
    for (var i = 0; i < 30 && pi != null && !pi.succeeded; i++) {
      await Future<void>.delayed(widget.pollInterval);
      if (!mounted) return;
      pi = await guarded(context, () => api.intent(pi!.id));
    }
    if (!mounted) return;
    if (pi != null && pi.succeeded) {
      final done = await guarded(context, () => api.booking(b.id));
      if (!mounted) return;
      setState(() {
        _busy = false;
        _booking = done ?? b;
        _step = _Step.success;
      });
    } else {
      setState(() => _busy = false);
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('ยังไม่ได้รับการยืนยันการชำระเงิน กรุณาตรวจสอบที่หน้า "การจอง"')));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(switch (_step) { _Step.details => 'เลือกวันที่และจำนวน', _Step.payment => 'ชำระเงิน', _Step.success => 'จองสำเร็จ' }),
        automaticallyImplyLeading: _step != _Step.success && !_busy,
      ),
      body: switch (_step) { _Step.details => _details(), _Step.payment => _payment(), _Step.success => _success() },
    );
  }

  Widget _summaryCard() => TepCard(
        child: Row(children: [
          ListingThumb(type: widget.listing.type, imageUrl: widget.listing.imageUrl, seed: widget.listing.id, size: 56),
          const SizedBox(width: 12),
          Expanded(
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(widget.listing.name, style: const TextStyle(fontWeight: FontWeight.w700)),
              Text(widget.listing.location, style: const TextStyle(color: TepColors.muted, fontSize: 13)),
            ]),
          ),
        ]),
      );

  Widget _details() => ListView(padding: const EdgeInsets.all(16), children: [
        _summaryCard(),
        const SizedBox(height: 16),
        TepCard(
          onTap: _pickDate,
          child: Row(children: [
            const Icon(Icons.calendar_month_rounded, color: TepColors.navyBar),
            const SizedBox(width: 12),
            const Expanded(child: Text('วันที่ใช้บริการ')),
            Text(dmy(_date), style: const TextStyle(fontWeight: FontWeight.w700)),
          ]),
        ),
        const SizedBox(height: 12),
        TepCard(
          child: Row(children: [
            const Icon(Icons.group_rounded, color: TepColors.navyBar),
            const SizedBox(width: 12),
            const Expanded(child: Text('จำนวน')),
            IconButton(key: const Key('qty-minus'), onPressed: _qty > 1 ? () => setState(() => _qty--) : null, icon: const Icon(Icons.remove_circle_outline)),
            Text('$_qty', key: const Key('qty'), style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 16)),
            IconButton(key: const Key('qty-plus'), onPressed: _qty < 20 ? () => setState(() => _qty++) : null, icon: const Icon(Icons.add_circle_outline)),
          ]),
        ),
        const SizedBox(height: 12),
        TextField(
          key: const Key('coupon'),
          controller: _coupon,
          textCapitalization: TextCapitalization.characters,
          decoration: const InputDecoration(labelText: 'รหัสคูปอง (ถ้ามี)', hintText: 'WELCOME10', prefixIcon: Icon(Icons.local_offer_outlined)),
        ),
        const SizedBox(height: 16),
        TepCard(
          child: Column(children: [
            _row('ราคา × $_qty', thb(_subtotal)),
            const Divider(),
            const Text('ค่าบริการและส่วนลดจะคำนวณตอนยืนยันการจอง', style: TextStyle(color: TepColors.muted, fontSize: 12)),
          ]),
        ),
        const SizedBox(height: 20),
        FilledButton(key: const Key('to-payment'), onPressed: _busy ? null : _createBooking, child: Text(_busy ? 'กำลังสร้างการจอง…' : 'ไปหน้าชำระเงิน')),
      ]);

  Widget _row(String l, String v, {bool bold = false, Color? color}) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 4),
        child: Row(children: [
          Expanded(child: Text(l, style: TextStyle(fontWeight: bold ? FontWeight.w800 : FontWeight.w400))),
          Text(v, style: TextStyle(fontWeight: bold ? FontWeight.w800 : FontWeight.w600, color: color)),
        ]),
      );

  Widget _payment() {
    final b = _booking!;
    return ListView(padding: const EdgeInsets.all(16), children: [
      TepCard(
        child: Column(children: [
          _row('เลขที่การจอง', b.bookingNumber),
          _row('ราคาสินค้า', thb(b.subtotal)),
          _row('ค่าบริการ', thb(b.serviceFee)),
          const Divider(),
          _row('ยอดชำระ', thb(b.amount), bold: true, color: TepColors.navyBar),
        ]),
      ),
      const SizedBox(height: 16),
      const Text('เลือกวิธีชำระเงิน', style: TextStyle(fontWeight: FontWeight.w700, fontSize: 16)),
      const SizedBox(height: 8),
      for (final m in const [('promptpay', 'PromptPay', Icons.qr_code_2_rounded), ('card', 'บัตรเครดิต/เดบิต', Icons.credit_card_rounded), ('wallet', 'Wallet', Icons.account_balance_wallet_rounded)])
        Padding(
          padding: const EdgeInsets.only(bottom: 8),
          child: TepCard(
            key: Key('method-${m.$1}'),
            onTap: _busy
                ? null
                : () => setState(() {
                      _method = m.$1;
                      _intentKey = newIdempotencyKey();
                    }),
            child: Row(children: [
              Icon(m.$3, color: TepColors.navyBar),
              const SizedBox(width: 12),
              Expanded(child: Text(m.$2, style: const TextStyle(fontWeight: FontWeight.w600))),
              Icon(_method == m.$1 ? Icons.radio_button_checked_rounded : Icons.radio_button_off_rounded, color: _method == m.$1 ? TepColors.navyBar : TepColors.muted),
            ]),
          ),
        ),
      if (_pending != null && !_pending!.succeeded) ...[
        const SizedBox(height: 8),
        TepCard(
          child: Column(children: [
            const SizedBox(height: 4),
            const CircularProgressIndicator(),
            const SizedBox(height: 12),
            Text(_pending!.method == 'promptpay' ? 'สแกน QR PromptPay เพื่อชำระ ${thb(_pending!.amount)}' : 'กำลังยืนยันการชำระเงิน…', textAlign: TextAlign.center),
            Text('อ้างอิง ${_pending!.id}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
          ]),
        ),
      ],
      const SizedBox(height: 16),
      FilledButton(key: const Key('pay'), onPressed: _busy ? null : _pay, child: Text(_busy ? 'กำลังดำเนินการ…' : 'ชำระเงิน ${thb(b.amount)}')),
    ]);
  }

  Widget _success() {
    final b = _booking!;
    return Container(
      width: double.infinity,
      color: TepColors.navy,
      padding: const EdgeInsets.all(24),
      child: SafeArea(
        child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
          Container(
            padding: const EdgeInsets.all(20),
            decoration: const BoxDecoration(color: TepColors.success, shape: BoxShape.circle),
            child: const Icon(Icons.check_rounded, color: Colors.white, size: 64),
          ),
          const SizedBox(height: 20),
          const Text('การจองสำเร็จ', style: TextStyle(color: Colors.white, fontSize: 26, fontWeight: FontWeight.w800)),
          const SizedBox(height: 8),
          Text(b.bookingNumber, key: const Key('booking-number'), style: const TextStyle(color: Colors.white70, fontSize: 16)),
          Text('${b.title} · ${thb(b.amount)}', style: const TextStyle(color: Colors.white70)),
          const SizedBox(height: 28),
          FilledButton(
            key: const Key('done'),
            style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: TepColors.navy),
            onPressed: () => Navigator.of(context).popUntil((r) => r.isFirst),
            child: const Text('กลับหน้าหลัก'),
          ),
        ]),
      ),
    );
  }
}
