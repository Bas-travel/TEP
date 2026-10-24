import 'dart:convert';

import 'package:http/http.dart' as http;

import 'api_client.dart';
import 'models.dart';

/// Real HTTP client for services/api-go. Tokens are kept in memory (secure storage is a Sprint 2 item).
class HttpApiClient implements ApiClient {
  HttpApiClient({required this.baseUrl, http.Client? client}) : _http = client ?? http.Client();

  final String baseUrl;
  final http.Client _http;
  AuthTokens? _tokens;

  Uri _uri(String path, [Map<String, String?> query = const {}]) {
    final q = <String, String>{for (final e in query.entries) if (e.value != null && e.value!.isNotEmpty) e.key: e.value!};
    return Uri.parse('$baseUrl/v1$path').replace(queryParameters: q.isEmpty ? null : q);
  }

  Future<Map<String, Object?>> _call(
    String method,
    Uri uri, {
    Object? body,
    Map<String, String> headers = const {},
    bool auth = true,
    bool retry = true,
  }) async {
    final h = <String, String>{'Accept': 'application/json', if (body != null) 'Content-Type': 'application/json', ...headers};
    if (auth && _tokens != null) h['Authorization'] = 'Bearer ${_tokens!.accessToken}';
    final req = http.Request(method, uri)..headers.addAll(h);
    if (body != null) req.body = jsonEncode(body);
    http.Response res;
    try {
      res = await http.Response.fromStream(await _http.send(req).timeout(const Duration(seconds: 15)));
    } catch (_) {
      throw const ApiException(status: 0, code: 'NETWORK', title: 'Network error');
    }
    if (res.statusCode == 401 && auth && retry && _tokens != null && await _refresh()) {
      return _call(method, uri, body: body, headers: headers, auth: auth, retry: false);
    }
    final text = utf8.decode(res.bodyBytes);
    Map<String, Object?> json = const {};
    if (text.trim().isNotEmpty) {
      try {
        json = (jsonDecode(text) as Map).cast<String, Object?>();
      } catch (_) {
        throw ApiException(status: res.statusCode, code: 'BAD_RESPONSE', title: 'Unexpected response');
      }
    }
    if (res.statusCode >= 400) throw ApiException.fromJson(res.statusCode, json);
    return json;
  }

  Future<bool> _refresh() async {
    final t = _tokens;
    if (t == null) return false;
    try {
      final j = await _call('POST', _uri('/auth/refresh'), body: {'refreshToken': t.refreshToken}, auth: false, retry: false);
      _tokens = AuthTokens.fromJson(j);
      return true;
    } on ApiException {
      _tokens = null;
      return false;
    }
  }

  List<Map<String, Object?>> _items(Map<String, Object?> j) =>
      ((j['items'] as List?) ?? const []).map((e) => (e as Map).cast<String, Object?>()).toList();

  @override
  Future<void> login(String email, String password) async {
    final j = await _call('POST', _uri('/auth/login'), body: {'email': email, 'password': password}, auth: false);
    _tokens = AuthTokens.fromJson(j);
  }

  @override
  Future<void> register({required String email, required String password, required String firstName, required String lastName}) async {
    final j = await _call('POST', _uri('/auth/register'),
        body: {'email': email, 'password': password, 'firstName': firstName, 'lastName': lastName}, auth: false);
    _tokens = AuthTokens.fromJson(j);
  }

  @override
  Future<void> logout() async {
    final t = _tokens;
    try {
      if (t != null) await _call('POST', _uri('/auth/logout'), body: {'refreshToken': t.refreshToken}, retry: false);
    } on ApiException {
      // Best effort: local session is cleared regardless.
    } finally {
      _tokens = null;
    }
  }

  @override
  Future<UserProfile> me() async => UserProfile.fromJson(await _call('GET', _uri('/me')));

  @override
  Future<List<Destination>> destinations() async =>
      _items(await _call('GET', _uri('/catalog/destinations'), auth: false)).map(Destination.fromJson).toList();

  @override
  Future<ListingPage> listings({String? q, String? type, int? destinationId, String? cursor, int limit = 20}) async {
    final j = await _call(
      'GET',
      _uri('/catalog/listings', {'q': q, 'type': type, 'destinationId': destinationId?.toString(), 'cursor': cursor, 'limit': '$limit'}),
      auth: false,
    );
    return ListingPage(_items(j).map(ListingSummary.fromJson).toList(), nextCursor: j['nextCursor'] as String?);
  }

  @override
  Future<ListingDetail> listing(int id) async => ListingDetail.fromJson(await _call('GET', _uri('/catalog/listings/$id'), auth: false));

  @override
  Future<List<Faq>> faqs() async => _items(await _call('GET', _uri('/support/faqs'), auth: false)).map(Faq.fromJson).toList();

  @override
  Future<List<BookingSummary>> bookings({String? status}) async =>
      _items(await _call('GET', _uri('/bookings', {'status': status}))).map(BookingSummary.fromJson).toList();

  @override
  Future<BookingDetail> booking(int id) async => BookingDetail.fromJson(await _call('GET', _uri('/bookings/$id')));

  @override
  Future<BookingDetail> createBooking({required List<BookingLine> lines, String? couponCode, required String idempotencyKey}) async {
    final j = await _call('POST', _uri('/bookings'),
        body: {'lines': lines.map((l) => l.toJson()).toList(), if (couponCode != null && couponCode.isNotEmpty) 'couponCode': couponCode},
        headers: {'Idempotency-Key': idempotencyKey});
    return BookingDetail.fromJson(j);
  }

  @override
  Future<BookingDetail> cancelBooking(int id, String reason) async =>
      BookingDetail.fromJson(await _call('POST', _uri('/bookings/$id/cancel'), body: {'reason': reason}));

  @override
  Future<PaymentIntent> createIntent({required int bookingId, required String method, required String idempotencyKey}) async =>
      PaymentIntent.fromJson(await _call('POST', _uri('/payments/intents'),
          body: {'bookingId': bookingId, 'method': method}, headers: {'Idempotency-Key': idempotencyKey}));

  @override
  Future<PaymentIntent> intent(String id) async => PaymentIntent.fromJson(await _call('GET', _uri('/payments/intents/$id')));

  @override
  Future<Wallet> wallet() async => Wallet.fromJson(await _call('GET', _uri('/wallet')));

  @override
  Future<PaymentIntent> topUp(double amount) async =>
      PaymentIntent.fromJson(await _call('POST', _uri('/wallet/top-ups'), body: {'amount': amount, 'method': 'promptpay'}));
}
