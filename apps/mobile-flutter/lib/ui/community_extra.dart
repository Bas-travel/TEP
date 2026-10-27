import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models_engage.dart';
import 'widgets/common.dart';

const kPostTypeLabels = {'story': 'เรื่องเล่า', 'tip': 'เคล็ดลับ', 'question': 'คำถาม'};

/// Bottom sheet to write a post (POST /community/posts). Pops with the created [Post], or null if dismissed.
class NewPostSheet extends StatefulWidget {
  const NewPostSheet({super.key, this.groupId});
  final int? groupId;
  @override
  State<NewPostSheet> createState() => _NewPostSheetState();
}

class _NewPostSheetState extends State<NewPostSheet> {
  final _text = TextEditingController();
  final _dest = TextEditingController();
  String _type = 'story';
  bool _busy = false;

  @override
  void dispose() {
    _text.dispose();
    _dest.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final api = AppScope.apiOf(context);
    setState(() => _busy = true);
    final p = await guarded(context, () => api.createPost(type: _type, destination: _dest.text.trim(), text: _text.text, groupId: widget.groupId));
    if (!mounted) return;
    setState(() => _busy = false);
    if (p != null) Navigator.of(context).pop(p);
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.fromLTRB(16, 16, 16, 16 + MediaQuery.of(context).viewInsets.bottom),
      child: SingleChildScrollView(
        child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
          const Text('เขียนโพสต์', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w800)),
          const SizedBox(height: 12),
          Wrap(spacing: 8, children: [
            for (final e in kPostTypeLabels.entries)
              ChoiceChip(key: Key('post-type-${e.key}'), label: Text(e.value), selected: _type == e.key, onSelected: (_) => setState(() => _type = e.key)),
          ]),
          const SizedBox(height: 12),
          TextField(key: const Key('post-dest'), controller: _dest, maxLength: 100, decoration: const InputDecoration(labelText: 'จุดหมาย (เช่น ภูเก็ต)')),
          TextField(
            key: const Key('post-text'),
            controller: _text,
            maxLines: 4,
            maxLength: 500,
            decoration: const InputDecoration(labelText: 'เล่าให้เพื่อนนักเดินทางฟัง (อย่างน้อย 10 ตัวอักษร)'),
          ),
          const SizedBox(height: 8),
          SizedBox(
            width: double.infinity,
            child: FilledButton(key: const Key('post-submit'), onPressed: _busy ? null : _submit, child: Text(_busy ? 'กำลังโพสต์…' : 'โพสต์')),
          ),
        ]),
      ),
    );
  }
}

/// Post with its comments and a comment box (GET /community/posts/{id}, POST …/comments).
class PostDetailScreen extends StatefulWidget {
  const PostDetailScreen({super.key, required this.postId});
  final int postId;
  @override
  State<PostDetailScreen> createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends State<PostDetailScreen> {
  final _text = TextEditingController();
  late Future<PostDetail> _f;
  bool _started = false;
  bool _busy = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_started) {
      _started = true;
      _f = AppScope.apiOf(context).postDetail(widget.postId);
    }
  }

  @override
  void dispose() {
    _text.dispose();
    super.dispose();
  }

  Future<void> _send() async {
    final api = AppScope.apiOf(context);
    setState(() => _busy = true);
    final c = await guarded(context, () => api.addComment(widget.postId, _text.text));
    if (!mounted) return;
    setState(() => _busy = false);
    if (c != null) {
      _text.clear();
      setState(() {
        _f = api.postDetail(widget.postId);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('โพสต์')),
      body: Column(children: [
        Expanded(
          child: FutureBuilder<PostDetail>(
            future: _f,
            builder: (context, snap) {
              if (snap.connectionState != ConnectionState.done) return const Center(child: CircularProgressIndicator());
              if (snap.hasError) return const EmptyState(Icons.cloud_off_rounded, 'โหลดโพสต์ไม่สำเร็จ');
              final d = snap.data!;
              return ListView(padding: const EdgeInsets.all(16), children: [
                TepCard(
                  child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text(d.post.displayName, style: const TextStyle(fontWeight: FontWeight.w700)),
                    Text('${kPostTypeLabels[d.post.type] ?? d.post.type} · ${d.post.destination} · ${dmy(d.post.createdAt)}', style: const TextStyle(color: TepColors.muted, fontSize: 12)),
                    const SizedBox(height: 10),
                    Text(d.post.text, key: const Key('detail-text')),
                  ]),
                ),
                SectionHeader('ความคิดเห็น (${d.comments.length})'),
                if (d.comments.isEmpty) const Padding(padding: EdgeInsets.all(16), child: Text('ยังไม่มีความคิดเห็น', style: TextStyle(color: TepColors.muted))),
                for (final c in d.comments)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: TepCard(
                      key: Key('comment-${c.id}'),
                      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                        Text(c.displayName, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 13)),
                        const SizedBox(height: 4),
                        Text(c.text),
                      ]),
                    ),
                  ),
              ]);
            },
          ),
        ),
        SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 8, 8),
            child: Row(children: [
              Expanded(child: TextField(key: const Key('comment-text'), controller: _text, maxLength: 1000, decoration: const InputDecoration(hintText: 'แสดงความคิดเห็น…', counterText: ''))),
              IconButton(key: const Key('comment-send'), onPressed: _busy ? null : _send, icon: const Icon(Icons.send_rounded, color: TepColors.primary)),
            ]),
          ),
        ),
      ]),
    );
  }
}

/// Travel groups (GET /community/groups, POST …/join toggles membership).
class GroupsScreen extends StatefulWidget {
  const GroupsScreen({super.key});
  @override
  State<GroupsScreen> createState() => _GroupsScreenState();
}

class _GroupsScreenState extends State<GroupsScreen> {
  int _version = 0;

  Future<void> _toggle(Group g) async {
    final api = AppScope.apiOf(context);
    final joined = await guarded(context, () => api.toggleGroup(g.id));
    if (joined != null && mounted) setState(() => _version++);
  }

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('กลุ่มนักเดินทาง')),
      body: AsyncBody<List<Group>>(
        key: ValueKey('groups-$_version'),
        load: api.groups,
        builder: (context, items, _) {
          if (items.isEmpty) return const EmptyState(Icons.groups_outlined, 'ยังไม่มีกลุ่ม');
          return ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: items.length,
            separatorBuilder: (_, __) => const SizedBox(height: 10),
            itemBuilder: (context, i) {
              final g = items[i];
              return TepCard(
                key: Key('group-${g.id}'),
                child: Row(children: [
                  Expanded(
                    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                      Text(g.name, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 16)),
                      Text(g.description, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted)),
                      const SizedBox(height: 4),
                      Text('${g.memberCount} สมาชิก', key: Key('group-members-${g.id}'), style: const TextStyle(fontSize: 12)),
                    ]),
                  ),
                  const SizedBox(width: 8),
                  g.joined
                      ? OutlinedButton(
                          key: Key('group-join-${g.id}'),
                          style: OutlinedButton.styleFrom(minimumSize: const Size(0, 40)),
                          onPressed: () => _toggle(g),
                          child: const Text('ออกจากกลุ่ม'))
                      : FilledButton(
                          key: Key('group-join-${g.id}'),
                          style: FilledButton.styleFrom(minimumSize: const Size(0, 40)),
                          onPressed: () => _toggle(g),
                          child: const Text('เข้าร่วม')),
                ]),
              );
            },
          );
        },
      ),
    );
  }
}

/// Change password (PUT /auth/password). The server revokes every session, so on success the user signs in again.
class ChangePasswordScreen extends StatefulWidget {
  const ChangePasswordScreen({super.key});
  @override
  State<ChangePasswordScreen> createState() => _ChangePasswordScreenState();
}

class _ChangePasswordScreenState extends State<ChangePasswordScreen> {
  final _cur = TextEditingController(), _new = TextEditingController();
  bool _busy = false;

  @override
  void dispose() {
    _cur.dispose();
    _new.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final session = AppScope.of(context);
    final navigator = Navigator.of(context);
    setState(() => _busy = true);
    final ok = await guarded(context, () async {
      await session.api.changePassword(currentPassword: _cur.text, newPassword: _new.text);
      return true;
    });
    if (!mounted) return;
    setState(() => _busy = false);
    if (ok == true) {
      navigator.popUntil((r) => r.isFirst);
      await session.logout();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('เปลี่ยนรหัสผ่าน')),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        TextField(key: const Key('pw-current'), controller: _cur, obscureText: true, decoration: const InputDecoration(labelText: 'รหัสผ่านปัจจุบัน')),
        const SizedBox(height: 12),
        TextField(key: const Key('pw-new'), controller: _new, obscureText: true, decoration: const InputDecoration(labelText: 'รหัสผ่านใหม่ (8–128 ตัวอักษร)')),
        const SizedBox(height: 8),
        const Text('เมื่อเปลี่ยนแล้ว ทุกอุปกรณ์จะถูกออกจากระบบ', style: TextStyle(color: TepColors.muted, fontSize: 12)),
        const SizedBox(height: 16),
        FilledButton(key: const Key('pw-submit'), onPressed: _busy ? null : _submit, child: Text(_busy ? 'กำลังบันทึก…' : 'เปลี่ยนรหัสผ่าน')),
      ]),
    );
  }
}
