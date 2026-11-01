import 'package:flutter/material.dart';

import '../app_scope.dart';
import '../core/format.dart';
import '../core/theme.dart';
import '../data/models_engage.dart';
import 'widgets/common.dart';

/// Renders an AI/community itinerary: per-day cards plus the budget breakdown.
class ItineraryView extends StatelessWidget {
  const ItineraryView(this.r, {super.key});
  final ItineraryResult r;

  @override
  Widget build(BuildContext context) {
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      for (final d in r.days)
        Padding(
          padding: const EdgeInsets.only(bottom: 10),
          child: TepCard(
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(d.title, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 15)),
              const SizedBox(height: 8),
              for (final it in d.items)
                Padding(
                  padding: const EdgeInsets.only(bottom: 6),
                  child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    SizedBox(width: 56, child: Text(it.time, maxLines: 1, style: const TextStyle(color: TepColors.partner, fontWeight: FontWeight.w700))),
                    Expanded(child: Text(it.text)),
                  ]),
                ),
            ]),
          ),
        ),
      if (r.budget.isNotEmpty)
        TepCard(
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            const Text('สรุปงบประมาณ', style: TextStyle(fontWeight: FontWeight.w800, fontSize: 15)),
            const SizedBox(height: 8),
            for (final b in r.budget)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 2),
                child: Row(children: [
                  Expanded(child: Text(b.label)),
                  Text(thb(b.value), style: const TextStyle(fontWeight: FontWeight.w700)),
                ]),
              ),
          ]),
        ),
    ]);
  }
}

/// Board 1.17 — My Trips (saved itineraries).
class TripsScreen extends StatefulWidget {
  const TripsScreen({super.key});
  @override
  State<TripsScreen> createState() => _TripsScreenState();
}

class _TripsScreenState extends State<TripsScreen> {
  int _version = 0;

  Future<void> _openPlanner() async {
    await Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => const AiPlannerScreen()));
    if (mounted) setState(() => _version++);
  }

  @override
  Widget build(BuildContext context) {
    final api = AppScope.apiOf(context);
    return Scaffold(
      appBar: AppBar(title: const Text('ทริปของฉัน')),
      floatingActionButton: FloatingActionButton.extended(
        key: const Key('plan-with-ai'),
        onPressed: _openPlanner,
        backgroundColor: TepColors.primary,
        foregroundColor: Colors.white,
        icon: const Icon(Icons.auto_awesome_rounded),
        label: const Text('วางแผนด้วย AI'),
      ),
      body: AsyncBody<List<SavedTrip>>(
        key: ValueKey('t-$_version'),
        load: api.savedTrips,
        builder: (context, items, reload) {
          if (items.isEmpty) return const EmptyState(Icons.luggage_outlined, 'ยังไม่มีทริปที่บันทึก\nให้ AI ช่วยวางแผนทริปแรกของคุณ');
          return RefreshIndicator(
            onRefresh: reload,
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 88),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (context, i) {
                final t = items[i];
                return TepCard(
                  key: Key('trip-${t.id}'),
                  onTap: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => TripDetailScreen(trip: t))),
                  child: Row(children: [
                    const Icon(Icons.luggage_rounded, color: TepColors.primary),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                        Text(t.destination, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16)),
                        Text('${t.days} วัน · ${t.travelers} คน · งบ ${thb(t.budget)}', maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(color: TepColors.muted, fontSize: 13)),
                      ]),
                    ),
                    const Icon(Icons.chevron_right_rounded),
                  ]),
                );
              },
            ),
          );
        },
      ),
    );
  }
}

class TripDetailScreen extends StatelessWidget {
  const TripDetailScreen({super.key, required this.trip});
  final SavedTrip trip;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(trip.destination, maxLines: 1, overflow: TextOverflow.ellipsis)),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        Text('${trip.days} วัน · ${trip.travelers} คน · งบ ${thb(trip.budget)}', style: const TextStyle(color: TepColors.muted)),
        const SizedBox(height: 12),
        if (trip.itinerary != null) ItineraryView(trip.itinerary!) else const EmptyState(Icons.map_outlined, 'ทริปนี้ยังไม่มีรายละเอียดแผน'),
      ]),
    );
  }
}

/// Board 1.27 — AI travel assistant (async job in the API; template fallback is flagged `degraded`).
class AiPlannerScreen extends StatefulWidget {
  const AiPlannerScreen({super.key});
  @override
  State<AiPlannerScreen> createState() => _AiPlannerScreenState();
}

class _AiPlannerScreenState extends State<AiPlannerScreen> {
  static const _styleOptions = ['ทะเล', 'คาเฟ่', 'วัฒนธรรม', 'ผจญภัย', 'อาหาร', 'ช้อปปิ้ง'];
  final _dest = TextEditingController(text: 'ภูเก็ต');
  final _budget = TextEditingController(text: '15000');
  final _styles = <String>{'ทะเล'};
  int _days = 3, _travelers = 2;
  bool _busy = false;
  ItineraryJob? _job;
  bool _saved = false;

  @override
  void dispose() {
    _dest.dispose();
    _budget.dispose();
    super.dispose();
  }

  ItineraryRequest _request() => ItineraryRequest(
        destination: _dest.text.trim(),
        days: _days,
        budget: double.tryParse(_budget.text.trim()) ?? 0,
        styles: _styles.toList(),
        travelers: _travelers,
      );

  Future<void> _generate() async {
    final api = AppScope.apiOf(context);
    setState(() {
      _busy = true;
      _saved = false;
    });
    final job = await guarded(context, () => api.planItinerary(_request()));
    if (!mounted) return;
    setState(() {
      _busy = false;
      if (job != null) _job = job;
    });
  }

  Future<void> _save() async {
    final api = AppScope.apiOf(context);
    final req = _request();
    final t = await guarded(
      context,
      () => api.saveTrip(destination: req.destination, days: req.days, travelers: req.travelers, budget: req.budget, itinerary: _job?.result),
    );
    if (!mounted || t == null) return;
    setState(() => _saved = true);
    ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('บันทึกทริปแล้ว')));
  }

  Widget _stepper(String label, int value, int min, int max, void Function(int) set, Key key) {
    return Row(children: [
      Expanded(child: Text(label, style: const TextStyle(fontWeight: FontWeight.w600))),
      IconButton(onPressed: value > min ? () => setState(() => set(value - 1)) : null, icon: const Icon(Icons.remove_circle_outline_rounded)),
      SizedBox(width: 28, child: Text('$value', key: key, textAlign: TextAlign.center, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16))),
      IconButton(onPressed: value < max ? () => setState(() => set(value + 1)) : null, icon: const Icon(Icons.add_circle_outline_rounded)),
    ]);
  }

  @override
  Widget build(BuildContext context) {
    final result = _job?.result;
    return Scaffold(
      appBar: AppBar(title: const Text('AI ผู้ช่วยวางแผน')),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        TepCard(
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            TextField(key: const Key('ai-destination'), controller: _dest, decoration: const InputDecoration(labelText: 'จุดหมาย', prefixIcon: Icon(Icons.place_outlined))),
            const SizedBox(height: 12),
            TextField(
              key: const Key('ai-budget'),
              controller: _budget,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(labelText: 'งบประมาณ (บาท)', prefixIcon: Icon(Icons.payments_outlined)),
            ),
            const SizedBox(height: 8),
            _stepper('จำนวนวัน', _days, 1, 14, (v) => _days = v, const Key('ai-days')),
            _stepper('ผู้เดินทาง', _travelers, 1, 10, (v) => _travelers = v, const Key('ai-travelers')),
            const SizedBox(height: 4),
            Wrap(spacing: 8, runSpacing: 4, children: [
              for (final s in _styleOptions)
                FilterChip(
                  key: Key('style-$s'),
                  label: Text(s),
                  selected: _styles.contains(s),
                  onSelected: (v) => setState(() => v ? _styles.add(s) : _styles.remove(s)),
                ),
            ]),
            const SizedBox(height: 16),
            FilledButton.icon(
              key: const Key('ai-generate'),
              onPressed: _busy ? null : _generate,
              icon: _busy ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white)) : const Icon(Icons.auto_awesome_rounded),
              label: const Text('สร้างแผนเที่ยว'),
            ),
          ]),
        ),
        if (result != null) ...[
          const SectionHeader('แผนที่แนะนำ'),
          if (_job!.degraded)
            const Padding(
              padding: EdgeInsets.only(bottom: 8),
              child: Text('สร้างจากเทมเพลตมาตรฐาน (AI จริงจะเชื่อมในเฟสถัดไป)', style: TextStyle(color: TepColors.muted, fontSize: 12)),
            ),
          ItineraryView(result),
          const SizedBox(height: 8),
          FilledButton.icon(
            key: const Key('ai-save'),
            onPressed: _saved ? null : _save,
            icon: Icon(_saved ? Icons.check_rounded : Icons.bookmark_add_outlined),
            label: Text(_saved ? 'บันทึกแล้ว' : 'บันทึกเป็นทริปของฉัน'),
          ),
        ],
      ]),
    );
  }
}
