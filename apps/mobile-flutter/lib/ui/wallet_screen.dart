import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'widgets/common.dart';

/// Board 1.19 — balance card, PromptPay top-up, transaction history.
class WalletScreen extends StatefulWidget {
  const WalletScreen({super.key});
  @override
  State<WalletScreen> createState() => _WalletScreenState();
}

class _WalletScreenState extends State<WalletScreen> {
  int _version = 0;

  Future<void> _topUp() async {
    final c = TextEditingController(text: '500');
    final amount = await showDialog<double>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('เติมเงินด้วย PromptPay'),
        content: TextField(key: const Key('topup-amount'), controller: c, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText: 'จำนวนเงิน (บาท)')),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('ยกเลิก')),
          TextButton(key: const Key('topup-confirm'), onPressed: () => Navigator.pop(ctx, double.tryParse(c.text.trim())), child: const Text('สร้าง QR')),
        ],
      ),
    );
    c.dispose();
    if (amount == null || !mounted) return;
    final api = AppScope.apiOf(context);
    final pi = await guarded(context, () => api.topUp(amount));
    if (pi == null || !mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('สร้างรายการเติมเงิน ${thb(pi.amount)} แล้ว (${pi.id}) — ยอดจะเข้าเมื่อชำระสำเร็จ')));
    setState(() => _version++);
  }

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('Wallet'), actions: [
        IconButton(tooltip: 'รีเฟรช', onPressed: () => setState(() => _version++), icon: const Icon(Icons.refresh_rounded)),
      ]),
      body: AsyncBody<Wallet>(
        key: ValueKey(_version),
        load: api.wallet,
        builder: (context, w, reload) => RefreshIndicator(
          onRefresh: reload,
          child: ListView(padding: const EdgeInsets.all(16), children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                gradient: const LinearGradient(colors: [TepColors.navy, TepColors.navyBar, TepColors.partner]),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                const Text('ยอดคงเหลือ', style: TextStyle(color: Colors.white70)),
                const SizedBox(height: 4),
                Text(thb(w.balance), key: const Key('wallet-balance'), style: const TextStyle(color: Colors.white, fontSize: 34, fontWeight: FontWeight.w800)),
                const SizedBox(height: 16),
                FilledButton.icon(
                  key: const Key('topup'),
                  style: FilledButton.styleFrom(backgroundColor: Colors.white, foregroundColor: TepColors.navy, minimumSize: const Size.fromHeight(46)),
                  onPressed: _topUp,
                  icon: const Icon(Icons.add_rounded),
                  label: const Text('เติมเงิน'),
                ),
              ]),
            ),
            const SectionHeader('รายการล่าสุด'),
            if (w.transactions.isEmpty)
              const Padding(padding: EdgeInsets.all(24), child: EmptyState(Icons.history_rounded, 'ยังไม่มีรายการ'))
            else
              for (final t in w.transactions)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: TepCard(
                    child: Row(children: [
                      CircleAvatar(
                        backgroundColor: (t.isIncoming ? TepColors.success : TepColors.danger).withValues(alpha: 0.12),
                        child: Icon(t.isIncoming ? Icons.south_west_rounded : Icons.north_east_rounded, color: t.isIncoming ? TepColors.success : TepColors.danger, size: 18),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                          Text(t.label, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w600)),
                          Text(dmy(t.createdAt), style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                        ]),
                      ),
                      Text('${t.isIncoming ? '+' : '-'}${thb(t.amount)}', style: TextStyle(fontWeight: FontWeight.w800, color: t.isIncoming ? TepColors.success : TepColors.danger)),
                    ]),
                  ),
                ),
          ]),
        ),
      ),
    );
  }
}
