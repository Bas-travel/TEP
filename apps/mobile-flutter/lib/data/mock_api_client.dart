import 'api_client.dart';
import 'mock_engage.dart';
import 'mock_fixtures.dart';
import 'models.dart';

/// In-memory implementation of the API contract so the app runs with no backend (API_MODE=mock).
/// Business rules mirror services/api-go mock mode: 3% service fee, WELCOME10 (10%, max 500, min 1000).
class MockApiClient with MockEngage implements ApiClient {
  MockApiClient({this.latency = const Duration(milliseconds: 250), this.settleAfter = const Duration(seconds: 3), DateTime Function()? now})
      : _now = now ?? DateTime.now {
    _wallet = 1500;
    _txs.add(WalletTx(id: _nextTx++, direction: 'in', label: 'เติมเงิน PromptPay', amount: 1500, createdAt: _now().subtract(const Duration(hours: 48))));
  }

  final Duration latency, settleAfter;
  final DateTime Function() _now;

  @override
  Duration get mockLatency => latency;
  @override
  bool get mockSignedIn => _signedIn;
  @override
  DateTime mockNow() => _now();

  static const demoEmail = 'demo@tep.dev';
  static const demoPassword = 'Passw0rd!';
  static const _methods = {'promptpay': 'PromptPay', 'card': 'บัตรเครดิต/เดบิต', 'wallet': 'Wallet'};

  bool _signedIn = false;
  UserProfile _user = const UserProfile(
    id: 1,
    email: demoEmail,
    firstName: 'สวัสดี',
    lastName: 'อนันต์',
    level: 'Explorer',
    favoriteDestination: 'ภูเก็ต',
    travelStyles: ['ทะเล', 'คาเฟ่'],
  );
  late double _wallet;
  final _txs = <WalletTx>[];
  final _bookings = <int, BookingDetail>{};
  final _intents = <String, _Intent>{};
  final _idem = <String, Object>{};
  int _nextBooking = 1, _nextTx = 1, _nextIntent = 1;

  Future<T> _t<T>(T Function() f) async {
    if (latency > Duration.zero) await Future<void>.delayed(latency);
    return f();
  }

  void _requireAuth() {
    if (!_signedIn) throw const ApiException(status: 401, code: 'AUTH_UNAUTHENTICATED', title: 'Unauthorized');
  }

  double _r2(double x) => (x * 100).roundToDouble() / 100;

  @override
  Future<void> login(String email, String password) => _t(() {
        if (email.trim().toLowerCase() != demoEmail || password != demoPassword) {
          throw const ApiException(status: 401, code: 'AUTH_UNAUTHENTICATED', title: 'Unauthorized', detail: 'invalid email or password');
        }
        _signedIn = true;
      });

  @override
  Future<void> register({required String email, required String password, required String firstName, required String lastName}) => _t(() {
        if (password.length < 8) {
          throw const ApiException(
              status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError('password', 'รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร')]);
        }
        _user = UserProfile(id: 2, email: email, firstName: firstName, lastName: lastName);
        _signedIn = true;
      });

  @override
  Future<void> logout() => _t(() => _signedIn = false);

  @override
  Future<UserProfile> me() => _t(() {
        _requireAuth();
        return _user;
      });

  @override
  Future<List<Destination>> destinations() => _t(() => kMockDestinations.map(Destination.fromJson).toList());

  @override
  Future<ListingPage> listings({String? q, String? type, int? destinationId, String? cursor, int limit = 20}) => _t(() {
        final needle = q?.trim().toLowerCase() ?? '';
        final all = kMockListings.where((l) {
          if (type != null && type.isNotEmpty && l['type'] != type) return false;
          if (destinationId != null && l['destinationId'] != destinationId) return false;
          if (needle.isNotEmpty && !'${l['name']} ${l['location']}'.toLowerCase().contains(needle)) return false;
          return true;
        }).map(ListingSummary.fromJson).toList()
          ..sort((a, b) => b.rating.compareTo(a.rating));
        final start = int.tryParse(cursor ?? '') ?? 0;
        if (start >= all.length) return const ListingPage([]);
        final end = start + limit;
        return ListingPage(all.sublist(start, end < all.length ? end : all.length), nextCursor: end < all.length ? '$end' : null);
      });

  @override
  Future<ListingDetail> listing(int id) => _t(() {
        final m = kMockListings.where((l) => l['id'] == id);
        if (m.isEmpty) throw const ApiException(status: 404, code: 'NOT_FOUND', title: 'Not Found');
        return ListingDetail.fromJson(m.first);
      });

  @override
  Future<List<Faq>> faqs() => _t(() => kMockFaqs.map(Faq.fromJson).toList());

  @override
  Future<List<BookingSummary>> bookings({String? status}) => _t(() {
        _requireAuth();
        final l = _bookings.values.where((b) => status == null || status == 'all' || status.isEmpty || b.status == status).toList()
          ..sort((a, b) => b.date.compareTo(a.date));
        return l;
      });

  @override
  Future<BookingDetail> booking(int id) => _t(() {
        _requireAuth();
        _settleAll();
        final b = _bookings[id];
        if (b == null) throw const ApiException(status: 404, code: 'NOT_FOUND', title: 'Not Found');
        return b;
      });

  @override
  Future<BookingDetail> createBooking({required List<BookingLine> lines, String? couponCode, required String idempotencyKey}) => _t(() {
        _requireAuth();
        final prior = _idem['b|$idempotencyKey'];
        if (prior is BookingDetail) return prior;
        if (lines.isEmpty) {
          throw const ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError('lines', 'at least one line required')]);
        }
        var subtotal = 0.0;
        ListingDetail? first;
        for (final ln in lines) {
          final m = kMockListings.where((l) => l['id'] == ln.listingId);
          if (m.isEmpty || ln.quantity < 1 || ln.quantity > 20) {
            throw const ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError('lines', 'unknown listing or bad quantity')]);
          }
          final d = ListingDetail.fromJson(m.first);
          subtotal += d.priceFrom * ln.quantity;
          first ??= d;
        }
        var discount = 0.0;
        final c = couponCode?.trim().toUpperCase() ?? '';
        if (c.isNotEmpty) {
          if (c != 'WELCOME10' || subtotal < 1000) {
            throw const ApiException(
                status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError('couponCode', 'คูปองไม่ถูกต้องหรือยอดไม่ถึงขั้นต่ำ')]);
          }
          discount = _r2(subtotal * 0.10).clamp(0, 500).toDouble();
        }
        final fee = _r2(subtotal * 0.03);
        final id = _nextBooking++;
        final now = _now();
        final b = BookingDetail(
          id: id,
          bookingNumber: 'TEP-${(now.year % 100).toString().padLeft(2, '0')}${now.month.toString().padLeft(2, '0')}${now.day.toString().padLeft(2, '0')}-${id.toString().padLeft(5, '0')}',
          type: first!.type,
          title: first.name,
          subtitle: first.location,
          amount: _r2(subtotal + fee - discount),
          method: '',
          date: now,
          status: 'pending_payment',
          subtotal: _r2(subtotal),
          serviceFee: fee,
        );
        _bookings[id] = b;
        _idem['b|$idempotencyKey'] = b;
        return b;
      });

  void _setBooking(BookingDetail b, {String? status, String? method, String? reason, double? refund}) {
    _bookings[b.id] = BookingDetail(
      id: b.id,
      bookingNumber: b.bookingNumber,
      type: b.type,
      title: b.title,
      subtitle: b.subtitle,
      amount: b.amount,
      method: method ?? b.method,
      date: b.date,
      status: status ?? b.status,
      subtotal: b.subtotal,
      serviceFee: b.serviceFee,
      cancelReason: reason ?? b.cancelReason,
      refundAmount: refund ?? b.refundAmount,
    );
  }

  @override
  Future<BookingDetail> cancelBooking(int id, String reason) => _t(() {
        _requireAuth();
        final b = _bookings[id];
        if (b == null) throw const ApiException(status: 404, code: 'NOT_FOUND', title: 'Not Found');
        if (reason.trim().isEmpty) {
          throw const ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError('reason', 'กรุณาระบุเหตุผล')]);
        }
        if (b.status != 'pending_payment' && b.status != 'confirmed') {
          throw ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError('status', 'ยกเลิกไม่ได้ในสถานะ ${b.status}')]);
        }
        var refund = 0.0;
        if (b.status == 'confirmed') {
          refund = b.amount;
          _wallet = _r2(_wallet + refund);
          _txs.add(WalletTx(id: _nextTx++, direction: 'in', label: 'คืนเงินการจอง ${b.bookingNumber}', amount: refund, createdAt: _now()));
        }
        _setBooking(b, status: 'cancelled', reason: reason, refund: refund);
        return _bookings[id]!;
      });

  @override
  Future<PaymentIntent> createIntent({required int bookingId, required String method, required String idempotencyKey}) => _t(() {
        _requireAuth();
        final prior = _idem['i|$idempotencyKey'];
        if (prior is PaymentIntent) return prior;
        final b = _bookings[bookingId];
        if (b == null) throw const ApiException(status: 404, code: 'NOT_FOUND', title: 'Not Found');
        if (!_methods.containsKey(method) || b.status != 'pending_payment') {
          throw const ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError('method', 'invalid method or booking state')]);
        }
        final id = 'pi_${(_nextIntent++).toString().padLeft(6, '0')}';
        late PaymentIntent out;
        if (method == 'wallet') {
          if (_wallet < b.amount) {
            throw const ApiException(status: 402, code: 'WALLET_INSUFFICIENT', title: 'Payment Required', detail: 'wallet balance is insufficient');
          }
          _wallet = _r2(_wallet - b.amount);
          _txs.add(WalletTx(id: _nextTx++, direction: 'out', label: 'ชำระการจอง ${b.bookingNumber}', amount: b.amount, createdAt: _now()));
          _setBooking(b, status: 'confirmed', method: _methods['wallet']);
          out = PaymentIntent(id: id, status: 'succeeded', method: method, amount: b.amount);
          _intents[id] = _Intent(out, _now(), settled: true);
        } else {
          out = PaymentIntent(
            id: id,
            status: 'pending',
            method: method,
            amount: b.amount,
            promptPayQrUrl: method == 'promptpay' ? 'https://mock.tep.local/qr/$id.png' : null,
            clientSecret: method == 'card' ? '${id}_secret_mock' : null,
          );
          _intents[id] = _Intent(out, _now(), bookingId: bookingId);
        }
        _idem['i|$idempotencyKey'] = out;
        return out;
      });

  void _settleAll() {
    for (final e in _intents.values) {
      if (e.settled || _now().difference(e.created) < settleAfter) continue;
      e.settled = true;
      e.intent = PaymentIntent(id: e.intent.id, status: 'succeeded', method: e.intent.method, amount: e.intent.amount);
      final bid = e.bookingId;
      if (bid != null) {
        final b = _bookings[bid];
        if (b != null && b.status == 'pending_payment') _setBooking(b, status: 'confirmed', method: _methods[e.intent.method]);
      } else {
        _wallet = _r2(_wallet + e.intent.amount);
        _txs.add(WalletTx(id: _nextTx++, direction: 'in', label: 'เติมเงิน PromptPay', amount: e.intent.amount, createdAt: _now()));
      }
    }
  }

  @override
  Future<PaymentIntent> intent(String id) => _t(() {
        _requireAuth();
        _settleAll();
        final e = _intents[id];
        if (e == null) throw const ApiException(status: 404, code: 'NOT_FOUND', title: 'Not Found');
        return e.intent;
      });

  @override
  Future<Wallet> wallet() => _t(() {
        _requireAuth();
        _settleAll();
        final txs = List<WalletTx>.of(_txs)..sort((a, b) => b.createdAt.compareTo(a.createdAt));
        return Wallet(balance: _wallet, currency: 'THB', transactions: txs);
      });

  @override
  Future<PaymentIntent> topUp(double amount) => _t(() {
        _requireAuth();
        if (amount < 1 || amount > 100000) {
          throw const ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError('amount', 'ระบุจำนวนเงิน 1 - 100,000 บาท')]);
        }
        final id = 'pi_${(_nextIntent++).toString().padLeft(6, '0')}';
        final pi = PaymentIntent(id: id, status: 'pending', method: 'promptpay', amount: _r2(amount), promptPayQrUrl: 'https://mock.tep.local/qr/$id.png');
        _intents[id] = _Intent(pi, _now());
        return pi;
      });
}

class _Intent {
  _Intent(this.intent, this.created, {this.bookingId, this.settled = false});
  PaymentIntent intent;
  final DateTime created;
  final int? bookingId;
  bool settled;
}
