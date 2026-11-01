import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models.dart';
import '../data/models_engage.dart';
import 'widgets/common.dart';

/// Board 1.23 (help center / FAQ) and 1.24 (contact us) in one screen with two tabs.
class SupportScreen extends StatelessWidget {
  const SupportScreen({super.key, this.initialTab = 0});
  final int initialTab;

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 2,
      initialIndex: initialTab,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('ช่วยเหลือ'),
          bottom: const TabBar(
            labelColor: Colors.white,
            unselectedLabelColor: Colors.white70,
            indicatorColor: Colors.white,
            tabs: [Tab(key: Key('tab-faq'), text: 'คำถามที่พบบ่อย'), Tab(key: Key('tab-contact'), text: 'ติดต่อเรา')],
          ),
        ),
        body: const TabBarView(children: [_FaqList(), _ContactTab()]),
      ),
    );
  }
}

class _FaqList extends StatelessWidget {
  const _FaqList();
  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return AsyncBody<List<Faq>>(
      load: api.faqs,
      builder: (context, items, _) => ListView(padding: const EdgeInsets.all(16), children: [
        for (final f in items)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: TepCard(
              padding: EdgeInsets.zero,
              child: ExpansionTile(
                shape: const Border(),
                collapsedShape: const Border(),
                title: Text(f.question, style: const TextStyle(fontWeight: FontWeight.w600)),
                childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                expandedAlignment: Alignment.centerLeft,
                children: [Text(f.answer)],
              ),
            ),
          ),
      ]),
    );
  }
}

class _ContactTab extends StatefulWidget {
  const _ContactTab();
  @override
  State<_ContactTab> createState() => _ContactTabState();
}

class _ContactTabState extends State<_ContactTab> {
  final _subject = TextEditingController();
  final _message = TextEditingController();
  bool _busy = false;
  int _version = 0;

  @override
  void dispose() {
    _subject.dispose();
    _message.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final api = AppScope.apiOf(context);
    setState(() => _busy = true);
    final t = await guarded(context, () => api.createTicket(subject: _subject.text.trim(), message: _message.text.trim()));
    if (!mounted) return;
    setState(() => _busy = false);
    if (t == null) return;
    _subject.clear();
    _message.clear();
    setState(() => _version++);
    ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('ส่งข้อความถึงทีมงานแล้ว')));
  }

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return ListView(padding: const EdgeInsets.all(16), children: [
      TepCard(
        child: Column(children: [
          TextField(key: const Key('ticket-subject'), controller: _subject, maxLength: 120, decoration: const InputDecoration(labelText: 'หัวข้อ')),
          const SizedBox(height: 8),
          TextField(key: const Key('ticket-message'), controller: _message, maxLines: 4, maxLength: 2000, decoration: const InputDecoration(labelText: 'รายละเอียด')),
          const SizedBox(height: 12),
          FilledButton(key: const Key('ticket-send'), onPressed: _busy ? null : _send, child: const Text('ส่งข้อความ')),
        ]),
      ),
      const SectionHeader('ข้อความที่ส่งแล้ว'),
      SizedBox(
        height: 260,
        child: AsyncBody<List<SupportTicket>>(
          key: ValueKey('tk-$_version'),
          load: api.tickets,
          builder: (context, items, _) {
            if (items.isEmpty) return const EmptyState(Icons.mark_email_unread_outlined, 'ยังไม่มีข้อความ');
            return ListView(children: [
              for (final t in items)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: TepCard(
                    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                      Text(t.subject, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700)),
                      Text(t.message, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted)),
                      Text(dmy(t.createdAt), style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                    ]),
                  ),
                ),
            ]);
          },
        ),
      ),
    ]);
  }
}
