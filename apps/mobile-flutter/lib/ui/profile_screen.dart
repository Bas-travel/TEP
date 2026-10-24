import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/theme.dart';
import '../data/models.dart';
import 'widgets/common.dart';

/// Board 1.20 — profile, plus 1.23 help center (FAQ) entry and logout.
class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final session = AppScope.of(context);
    final u = session.user;
    return Scaffold(
      appBar: AppBar(title: const Text('โปรไฟล์')),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        TepCard(
          padding: const EdgeInsets.all(20),
          child: Column(children: [
            CircleAvatar(
              radius: 40,
              backgroundColor: TepColors.navyBar,
              child: Text((u?.firstName.isNotEmpty ?? false) ? u!.firstName.characters.first : '?', style: const TextStyle(color: Colors.white, fontSize: 32)),
            ),
            const SizedBox(height: 12),
            Text(u?.fullName ?? '', key: const Key('profile-name'), style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w800)),
            Text(u?.email ?? '', style: const TextStyle(color: TepColors.muted)),
            const SizedBox(height: 8),
            Chip(label: Text(u?.level ?? ''), avatar: const Icon(Icons.workspace_premium_rounded, size: 18, color: TepColors.accent)),
            if (u != null && u.travelStyles.isNotEmpty) ...[
              const SizedBox(height: 8),
              Wrap(spacing: 6, children: [for (final s in u.travelStyles) Chip(label: Text(s), visualDensity: VisualDensity.compact)]),
            ],
          ]),
        ),
        const SizedBox(height: 16),
        TepCard(
          padding: EdgeInsets.zero,
          child: Column(children: [
            ListTile(
              key: const Key('open-faq'),
              leading: const Icon(Icons.help_outline_rounded, color: TepColors.navyBar),
              title: const Text('ศูนย์ช่วยเหลือ'),
              trailing: const Icon(Icons.chevron_right_rounded),
              onTap: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const FaqScreen())),
            ),
            const Divider(height: 1),
            ListTile(
              key: const Key('logout'),
              leading: const Icon(Icons.logout_rounded, color: TepColors.danger),
              title: const Text('ออกจากระบบ', style: TextStyle(color: TepColors.danger)),
              onTap: () => guarded(context, session.logout),
            ),
          ]),
        ),
      ]),
    );
  }
}

class FaqScreen extends StatelessWidget {
  const FaqScreen({super.key});
  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('ศูนย์ช่วยเหลือ')),
      body: AsyncBody<List<Faq>>(
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
      ),
    );
  }
}
