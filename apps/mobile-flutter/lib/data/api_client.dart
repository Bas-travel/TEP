import 'models.dart';

class BookingLine {
  const BookingLine(this.listingId, this.quantity);
  final int listingId, quantity;
  Map<String, Object?> toJson() => {'listingId': listingId, 'quantity': quantity};
}

/// Single seam between UI and backend. `MockApiClient` (API_MODE=mock) and
/// `HttpApiClient` (API_MODE=http) implement the same contract (docs/api/openapi.yaml).
abstract class ApiClient {
  Future<void> login(String email, String password);
  Future<void> register({required String email, required String password, required String firstName, required String lastName});
  Future<void> logout();
  Future<UserProfile> me();

  Future<List<Destination>> destinations();
  Future<ListingPage> listings({String? q, String? type, int? destinationId, String? cursor, int limit = 20});
  Future<ListingDetail> listing(int id);
  Future<List<Faq>> faqs();

  Future<List<BookingSummary>> bookings({String? status});
  Future<BookingDetail> booking(int id);
  Future<BookingDetail> createBooking({required List<BookingLine> lines, String? couponCode, required String idempotencyKey});
  Future<BookingDetail> cancelBooking(int id, String reason);

  Future<PaymentIntent> createIntent({required int bookingId, required String method, required String idempotencyKey});
  Future<PaymentIntent> intent(String id);
  Future<Wallet> wallet();
  Future<PaymentIntent> topUp(double amount);
}
