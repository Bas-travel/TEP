/// Wire models. Field names follow docs/api/openapi.yaml (camelCase).
double _d(Object? v) => (v as num?)?.toDouble() ?? 0;
int _i(Object? v) => (v as num?)?.toInt() ?? 0;
List<String> _ls(Object? v) => (v as List?)?.map((e) => e.toString()).toList() ?? const [];

class Destination {
  const Destination({required this.id, required this.slug, required this.name, required this.subtitle, required this.gradientCss});
  final int id;
  final String slug, name, subtitle, gradientCss;
  factory Destination.fromJson(Map<String, Object?> j) => Destination(
        id: _i(j['id']),
        slug: j['slug'] as String? ?? '',
        name: j['name'] as String? ?? '',
        subtitle: j['subtitle'] as String? ?? '',
        gradientCss: j['gradientCss'] as String? ?? '',
      );
}

class ListingSummary {
  const ListingSummary({
    required this.id,
    required this.type,
    required this.name,
    required this.location,
    required this.rating,
    required this.reviewCount,
    required this.priceFrom,
    this.imageUrl,
  });
  final int id, reviewCount;
  final String type, name, location;
  final double rating, priceFrom;
  final String? imageUrl;
  factory ListingSummary.fromJson(Map<String, Object?> j) => ListingSummary(
        id: _i(j['id']),
        type: j['type'] as String? ?? '',
        name: j['name'] as String? ?? '',
        location: j['location'] as String? ?? '',
        rating: _d(j['rating']),
        reviewCount: _i(j['reviewCount']),
        priceFrom: _d(j['priceFrom']),
        imageUrl: j['imageUrl'] as String?,
      );
}

class ListingDetail extends ListingSummary {
  const ListingDetail({
    required super.id,
    required super.type,
    required super.name,
    required super.location,
    required super.rating,
    required super.reviewCount,
    required super.priceFrom,
    super.imageUrl,
    required this.description,
    required this.amenities,
    required this.mapQuery,
  });
  final String description, mapQuery;
  final List<String> amenities;
  factory ListingDetail.fromJson(Map<String, Object?> j) {
    final s = ListingSummary.fromJson(j);
    return ListingDetail(
      id: s.id,
      type: s.type,
      name: s.name,
      location: s.location,
      rating: s.rating,
      reviewCount: s.reviewCount,
      priceFrom: s.priceFrom,
      imageUrl: s.imageUrl,
      description: j['description'] as String? ?? '',
      amenities: _ls(j['amenities']),
      mapQuery: j['mapQuery'] as String? ?? '',
    );
  }
}

class ListingPage {
  const ListingPage(this.items, {this.nextCursor});
  final List<ListingSummary> items;
  final String? nextCursor;
}

class Faq {
  const Faq(this.id, this.question, this.answer);
  final int id;
  final String question, answer;
  factory Faq.fromJson(Map<String, Object?> j) => Faq(_i(j['id']), j['question'] as String? ?? '', j['answer'] as String? ?? '');
}

class AuthTokens {
  const AuthTokens({required this.accessToken, required this.refreshToken, required this.expiresIn});
  final String accessToken, refreshToken;
  final int expiresIn;
  factory AuthTokens.fromJson(Map<String, Object?> j) => AuthTokens(
        accessToken: j['accessToken'] as String? ?? '',
        refreshToken: j['refreshToken'] as String? ?? '',
        expiresIn: _i(j['expiresIn']),
      );
}

class UserProfile {
  const UserProfile({
    required this.id,
    required this.email,
    required this.firstName,
    required this.lastName,
    this.level = 'Explorer',
    this.favoriteDestination,
    this.travelStyles = const [],
    this.avatarUrl,
  });
  final int id;
  final String email, firstName, lastName, level;
  final String? favoriteDestination, avatarUrl;
  final List<String> travelStyles;
  String get fullName => '$firstName $lastName'.trim();
  factory UserProfile.fromJson(Map<String, Object?> j) => UserProfile(
        id: _i(j['id']),
        email: j['email'] as String? ?? '',
        firstName: j['firstName'] as String? ?? '',
        lastName: j['lastName'] as String? ?? '',
        level: j['level'] as String? ?? 'Explorer',
        favoriteDestination: j['favoriteDestination'] as String?,
        travelStyles: _ls(j['travelStyles']),
        avatarUrl: j['avatarUrl'] as String?,
      );
}

class BookingSummary {
  const BookingSummary({
    required this.id,
    required this.bookingNumber,
    required this.type,
    required this.title,
    required this.subtitle,
    required this.amount,
    required this.method,
    required this.date,
    required this.status,
    this.startDate,
    this.endDate,
    this.guests = 1,
  });
  final int id;
  final String bookingNumber, type, title, subtitle, method, status;

  /// Stay / visit dates (calendar dates, `YYYY-MM-DD` on the wire) and number of travellers — finding F-1.
  final String? startDate, endDate;
  final int guests;
  final double amount;
  final DateTime date;
  factory BookingSummary.fromJson(Map<String, Object?> j) => BookingSummary(
        id: _i(j['id']),
        bookingNumber: j['bookingNumber'] as String? ?? '',
        type: j['type'] as String? ?? '',
        title: j['title'] as String? ?? '',
        subtitle: j['subtitle'] as String? ?? '',
        amount: _d(j['amount']),
        method: j['method'] as String? ?? '',
        date: DateTime.tryParse(j['date'] as String? ?? '') ?? DateTime.fromMillisecondsSinceEpoch(0),
        status: j['status'] as String? ?? '',
        startDate: j['startDate'] as String?,
        endDate: j['endDate'] as String?,
        guests: (j['guests'] as num?)?.toInt() ?? 1,
      );
}

class BookingDetail extends BookingSummary {
  const BookingDetail({
    required super.id,
    required super.bookingNumber,
    required super.type,
    required super.title,
    required super.subtitle,
    required super.amount,
    required super.method,
    required super.date,
    required super.status,
    super.startDate,
    super.endDate,
    super.guests,
    required this.subtotal,
    required this.serviceFee,
    this.cancelReason,
    this.refundAmount,
  });
  final double subtotal, serviceFee;
  final String? cancelReason;
  final double? refundAmount;
  factory BookingDetail.fromJson(Map<String, Object?> j) {
    final s = BookingSummary.fromJson(j);
    return BookingDetail(
      id: s.id,
      bookingNumber: s.bookingNumber,
      type: s.type,
      title: s.title,
      subtitle: s.subtitle,
      amount: s.amount,
      method: s.method,
      date: s.date,
      status: s.status,
      startDate: s.startDate,
      endDate: s.endDate,
      guests: s.guests,
      subtotal: _d(j['subtotal']),
      serviceFee: _d(j['serviceFee']),
      cancelReason: j['cancelReason'] as String?,
      refundAmount: (j['refundAmount'] as num?)?.toDouble(),
    );
  }
}

class PaymentIntent {
  const PaymentIntent({required this.id, required this.status, required this.method, required this.amount, this.promptPayQrUrl, this.clientSecret});
  final String id, status, method;
  final double amount;
  final String? promptPayQrUrl, clientSecret;
  bool get succeeded => status == 'succeeded';
  factory PaymentIntent.fromJson(Map<String, Object?> j) => PaymentIntent(
        id: j['id'] as String? ?? '',
        status: j['status'] as String? ?? '',
        method: j['method'] as String? ?? '',
        amount: _d(j['amount']),
        promptPayQrUrl: j['promptPayQrUrl'] as String?,
        clientSecret: j['clientSecret'] as String?,
      );
}

class WalletTx {
  const WalletTx({required this.id, required this.direction, required this.label, required this.amount, required this.createdAt});
  final int id;
  final String direction, label;
  final double amount;
  final DateTime createdAt;
  bool get isIncoming => direction == 'in';
  factory WalletTx.fromJson(Map<String, Object?> j) => WalletTx(
        id: _i(j['id']),
        direction: j['direction'] as String? ?? 'in',
        label: j['label'] as String? ?? '',
        amount: _d(j['amount']),
        createdAt: DateTime.tryParse(j['createdAt'] as String? ?? '') ?? DateTime.fromMillisecondsSinceEpoch(0),
      );
}

class Wallet {
  const Wallet({required this.balance, required this.currency, required this.transactions});
  final double balance;
  final String currency;
  final List<WalletTx> transactions;
  factory Wallet.fromJson(Map<String, Object?> j) => Wallet(
        balance: _d(j['balance']),
        currency: j['currency'] as String? ?? 'THB',
        transactions: ((j['transactions'] as List?) ?? const [])
            .map((e) => WalletTx.fromJson((e as Map).cast<String, Object?>()))
            .toList(),
      );
}

class FieldError {
  const FieldError(this.field, this.message);
  final String field, message;
}

/// RFC 7807 problem as returned by the API (docs/api/conventions.md).
class ApiException implements Exception {
  const ApiException({required this.status, required this.code, required this.title, this.detail, this.errors = const [], this.requestId});
  final int status;
  final String code, title;
  final String? detail, requestId;
  final List<FieldError> errors;

  bool get isUnauthenticated => status == 401;
  bool get isWalletShort => code == 'WALLET_INSUFFICIENT';

  /// Text safe to show to the user (Thai UI copy).
  String get userMessage {
    if (errors.isNotEmpty) return errors.map((e) => e.message).join('\n');
    if (isWalletShort) return 'ยอดเงินใน Wallet ไม่เพียงพอ';
    if (status == 401) return 'อีเมลหรือรหัสผ่านไม่ถูกต้อง หรือเซสชันหมดอายุ';
    if (status == 0) return 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาลองใหม่';
    return detail ?? title;
  }

  factory ApiException.fromJson(int status, Map<String, Object?> j) => ApiException(
        status: status,
        code: j['code'] as String? ?? 'UNKNOWN',
        title: j['title'] as String? ?? 'Error',
        detail: j['detail'] as String?,
        requestId: j['requestId'] as String?,
        errors: ((j['errors'] as List?) ?? const [])
            .map((e) => (e as Map).cast<String, Object?>())
            .map((e) => FieldError(e['field'] as String? ?? '', e['message'] as String? ?? ''))
            .toList(),
      );

  @override
  String toString() => 'ApiException($status $code: $title)';
}
