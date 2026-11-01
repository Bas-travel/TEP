import 'models.dart';
import 'models_engage.dart';

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
  Future<BookingDetail> createBooking({
    required List<BookingLine> lines,
    String? couponCode,
    DateTime? startDate,
    DateTime? endDate,
    int guests = 1,
    required String idempotencyKey,
  });
  Future<BookingDetail> cancelBooking(int id, String reason);

  Future<PaymentIntent> createIntent({required int bookingId, required String method, required String idempotencyKey});
  Future<PaymentIntent> intent(String id);
  Future<Wallet> wallet();
  Future<PaymentIntent> topUp(double amount);

  // ---- engagement (Sprint 2): settings, coupons, notifications, trips, reviews, support, community, AI ----
  Future<UserSettings> settings();
  Future<UserSettings> updateSettings(UserSettings s);

  Future<List<Coupon>> coupons();
  Future<double> applyCoupon(String code, double bookingTotal);

  Future<List<AppNotification>> notifications();
  Future<void> markNotificationRead(int id);
  Future<void> markAllNotificationsRead();

  Future<List<SavedTrip>> savedTrips();
  Future<SavedTrip> saveTrip({required String destination, required int days, required int travelers, required double budget, ItineraryResult? itinerary, String source = 'ai'});
  Future<ItineraryJob> planItinerary(ItineraryRequest req);

  Future<List<Review>> myReviews();
  Future<Review> createReview({required int listingId, required int rating, required String comment});

  Future<List<SupportTicket>> tickets();
  Future<SupportTicket> createTicket({required String subject, required String message});

  Future<List<Post>> feed({String mode = 'all', String? destination});
  Future<LikeResult> toggleLike(int postId);
  Future<bool> toggleSave(int postId);
  Future<Post> createPost({required String type, required String destination, required String text, int? groupId});
  Future<PostDetail> postDetail(int postId);
  Future<Comment> addComment(int postId, String text);
  Future<List<Group>> groups();
  Future<Group> group(int id);

  /// Returns true when the caller is now a member.
  Future<bool> toggleGroup(int id);

  // ---- account ----
  Future<void> changePassword({required String currentPassword, required String newPassword});
  Future<void> deleteAccount();
}
