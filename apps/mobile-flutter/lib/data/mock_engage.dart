import 'api_client.dart';
import 'mock_fixtures.dart';
import 'models.dart';
import 'models_engage.dart';

/// Mock implementation of the engagement endpoints. Mixed into `MockApiClient`
/// (API_MODE=mock); behaviour mirrors services/api-go `handlers_engage.go`.
mixin MockEngage implements ApiClient {
  Duration get mockLatency;
  bool get mockSignedIn;
  DateTime mockNow();

  UserSettings _settings = const UserSettings();
  final List<Coupon> _coupons = [
    const Coupon(id: 1, code: 'WELCOME10', title: 'ส่วนลดสมาชิกใหม่ 10%', detail: 'ลดสูงสุด ฿500 เมื่อใช้จ่ายขั้นต่ำ ฿1,000', minSpend: 1000, maxDiscount: 500, expiry: '2026-12-31', status: 'available'),
    const Coupon(id: 2, code: 'SEA200', title: 'ลด ฿200 ทริปทะเล', detail: 'เมื่อใช้จ่ายขั้นต่ำ ฿2,000', minSpend: 2000, maxDiscount: 200, expiry: '2026-11-30', status: 'available'),
    const Coupon(id: 3, code: 'OLD50', title: 'คูปองหมดอายุ', detail: 'ลด ฿50', minSpend: 500, maxDiscount: 50, expiry: '2025-12-31', status: 'expired'),
  ];
  late final List<AppNotification> _notifs = [
    AppNotification(id: 1, type: 'promo', title: 'คูปองต้อนรับ', message: 'รับส่วนลด 10% สำหรับการจองแรก', read: false, createdAt: mockNow().subtract(const Duration(hours: 2))),
    AppNotification(id: 2, type: 'trip', title: 'ทริปใกล้ถึงแล้ว', message: 'เตรียมตัวสำหรับทริปภูเก็ตของคุณ', read: false, createdAt: mockNow().subtract(const Duration(hours: 20))),
    AppNotification(id: 3, type: 'booking', title: 'การจองสำเร็จ', message: 'การจองของคุณได้รับการยืนยันแล้ว', read: true, createdAt: mockNow().subtract(const Duration(days: 2))),
    AppNotification(id: 4, type: 'system', title: 'ยินดีต้อนรับสู่ TEP', message: 'เริ่มวางแผนทริปแรกของคุณได้เลย', read: true, createdAt: mockNow().subtract(const Duration(days: 5))),
  ];
  final List<SavedTrip> _trips = [];
  final List<Review> _reviews = [];
  final List<SupportTicket> _tickets = [];
  late final List<Post> _posts = [
    Post(id: 1, authorId: 'm_nok', type: 'story', destination: 'ภูเก็ต', text: 'พระอาทิตย์ตกที่หาดกะตะสวยมาก แนะนำให้มาก่อน 17:30', likeCount: 42, hasImage: true, createdAt: mockNow().subtract(const Duration(hours: 3))),
    Post(id: 2, authorId: 'm_pim', type: 'tip', destination: 'เชียงใหม่', text: 'ขึ้นดอยสุเทพช่วงเช้า อากาศเย็นและคนน้อย', likeCount: 18, hasImage: false, createdAt: mockNow().subtract(const Duration(hours: 9))),
    Post(id: 3, authorId: 'm_art', type: 'question', destination: 'กระบี่', text: 'ใครเคยไปไร่เลย์ช่วงหน้าฝนบ้าง เรือออกไหม?', likeCount: 5, hasImage: false, createdAt: mockNow().subtract(const Duration(days: 1))),
    Post(id: 4, authorId: 'm_nok', type: 'trip', destination: 'เกาะสมุย', text: 'แพลน 3 วัน 2 คืน งบ 8,000 แชร์ให้ครับ', likeCount: 27, hasImage: true, createdAt: mockNow().subtract(const Duration(days: 2)), sharedDays: 3),
    Post(id: 5, authorId: 'm_lek', type: 'story', destination: 'กรุงเทพฯ', text: 'สตรีทฟู้ดประตูผีคุ้มค่ามาก ไปตอนค่ำ', likeCount: 61, hasImage: true, createdAt: mockNow().subtract(const Duration(days: 3))),
  ];
  final Set<int> _liked = {};
  int _nextTrip = 1, _nextReview = 1, _nextTicket = 1;

  Future<T> _e<T>(T Function() f) async {
    if (mockLatency > Duration.zero) await Future<void>.delayed(mockLatency);
    if (!mockSignedIn) throw const ApiException(status: 401, code: 'AUTH_UNAUTHENTICATED', title: 'Unauthorized');
    return f();
  }

  ApiException _invalid(String field, String msg) =>
      ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError(field, msg)]);

  @override
  Future<UserSettings> settings() => _e(() => _settings);

  @override
  Future<UserSettings> updateSettings(UserSettings s) => _e(() {
        if (!const {'light', 'dark', 'system'}.contains(s.theme)) throw _invalid('theme', 'ธีมไม่ถูกต้อง');
        return _settings = s;
      });

  @override
  Future<List<Coupon>> coupons() => _e(() => List.of(_coupons));

  @override
  Future<double> applyCoupon(String code, double bookingTotal) => _e(() {
        final c = _coupons.where((e) => e.code == code.toUpperCase());
        if (c.isEmpty) throw const ApiException(status: 404, code: 'NOT_FOUND', title: 'Not Found');
        final cp = c.first;
        if (!cp.usable) throw _invalid('code', 'คูปองนี้ใช้ไม่ได้');
        if (bookingTotal < cp.minSpend) throw _invalid('code', 'ยอดขั้นต่ำ ${cp.minSpend.toStringAsFixed(0)} บาท');
        final pct = cp.code == 'WELCOME10' ? bookingTotal * 0.10 : cp.maxDiscount;
        return pct > cp.maxDiscount ? cp.maxDiscount : pct;
      });

  @override
  Future<List<AppNotification>> notifications() => _e(() => List.of(_notifs)..sort((a, b) => b.createdAt.compareTo(a.createdAt)));

  @override
  Future<void> markNotificationRead(int id) => _e(() {
        final i = _notifs.indexWhere((n) => n.id == id);
        if (i < 0) throw const ApiException(status: 404, code: 'NOT_FOUND', title: 'Not Found');
        _notifs[i] = _notifs[i].asRead();
      });

  @override
  Future<void> markAllNotificationsRead() => _e(() {
        for (var i = 0; i < _notifs.length; i++) {
          _notifs[i] = _notifs[i].asRead();
        }
      });

  @override
  Future<List<SavedTrip>> savedTrips() => _e(() => List.of(_trips));

  @override
  Future<SavedTrip> saveTrip({required String destination, required int days, required int travelers, required double budget, ItineraryResult? itinerary, String source = 'ai'}) => _e(() {
        if (destination.trim().isEmpty) throw _invalid('destination', 'กรุณาระบุจุดหมาย');
        if (days < 1 || days > 14) throw _invalid('days', 'จำนวนวันต้องอยู่ระหว่าง 1–14');
        final t = SavedTrip(id: _nextTrip++, destination: destination, days: days, travelers: travelers, budget: budget, source: source, itinerary: itinerary);
        _trips.insert(0, t);
        return t;
      });

  @override
  Future<ItineraryJob> planItinerary(ItineraryRequest req) => _e(() {
        if (req.destination.trim().isEmpty) throw _invalid('destination', 'กรุณาระบุจุดหมาย');
        if (req.days < 1 || req.days > 14) throw _invalid('days', 'จำนวนวันต้องอยู่ระหว่าง 1–14');
        if (req.budget <= 0) throw _invalid('budget', 'งบประมาณต้องมากกว่า 0');
        const slots = [
          ['09:00', 'เช้า: เที่ยวชมจุดเด่น'],
          ['13:00', 'กลางวัน: อาหารท้องถิ่น'],
          ['16:00', 'บ่าย: คาเฟ่และช้อปปิ้ง'],
          ['19:00', 'เย็น: ชมวิวและอาหารค่ำ'],
        ];
        final days = [
          for (var d = 1; d <= req.days; d++)
            ItineraryDay(title: 'วันที่ $d · ${req.destination}', items: [for (final s in slots) ItineraryItem(time: s[0], text: s[1], place: req.destination)]),
        ];
        final b = req.budget;
        return ItineraryJob(
          jobId: 'job_mock',
          status: 'completed',
          degraded: true,
          result: ItineraryResult(days: days, budget: [BudgetLine('ที่พัก', b * 0.4), BudgetLine('อาหาร', b * 0.25), BudgetLine('เดินทาง', b * 0.2), BudgetLine('กิจกรรม', b * 0.15)]),
        );
      });

  @override
  Future<List<Review>> myReviews() => _e(() => List.of(_reviews));

  @override
  Future<Review> createReview({required int listingId, required int rating, required String comment}) => _e(() {
        if (rating < 1 || rating > 5) throw _invalid('rating', 'คะแนนต้องอยู่ระหว่าง 1–5');
        if (comment.trim().isEmpty) throw _invalid('comment', 'กรุณาเขียนความคิดเห็น');
        final l = kMockListings.where((e) => e['id'] == listingId);
        if (l.isEmpty) throw _invalid('listingId', 'ไม่พบรายการ');
        final r = Review(id: _nextReview++, itemName: l.first['name'] as String, itemType: l.first['type'] as String, rating: rating, comment: comment, createdAt: mockNow());
        _reviews.insert(0, r);
        return r;
      });

  @override
  Future<List<SupportTicket>> tickets() => _e(() => List.of(_tickets));

  @override
  Future<SupportTicket> createTicket({required String subject, required String message}) => _e(() {
        if (subject.trim().isEmpty) throw _invalid('subject', 'กรุณาระบุหัวข้อ');
        if (subject.length > 120) throw _invalid('subject', 'หัวข้อยาวเกิน 120 ตัวอักษร');
        if (message.trim().isEmpty) throw _invalid('message', 'กรุณาระบุข้อความ');
        final t = SupportTicket(id: _nextTicket++, subject: subject, message: message, createdAt: mockNow());
        _tickets.insert(0, t);
        return t;
      });

  @override
  Future<List<Post>> feed({String mode = 'all', String? destination}) => _e(() {
        Iterable<Post> l = _posts;
        if (mode == 'following') l = l.where((p) => p.authorId == 'm_nok' || p.authorId == 'm_pim');
        if (destination != null && destination.isNotEmpty) l = l.where((p) => p.destination == destination);
        return l.map((p) => p.withLike(_liked.contains(p.id), p.likeCount)).toList();
      });

  @override
  Future<LikeResult> toggleLike(int postId) => _e(() {
        final i = _posts.indexWhere((p) => p.id == postId);
        if (i < 0) throw const ApiException(status: 404, code: 'NOT_FOUND', title: 'Not Found');
        final liked = _liked.add(postId);
        if (!liked) _liked.remove(postId);
        final p = _posts[i];
        final count = p.likeCount + (liked ? 1 : -1);
        _posts[i] = p.withLike(liked, count);
        return LikeResult(liked, count);
      });
}
