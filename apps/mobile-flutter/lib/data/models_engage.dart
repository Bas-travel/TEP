/// Wire models for the engagement endpoints (settings, coupons, notifications, trips,
/// reviews, support, community, AI planner). Field names follow docs/api/openapi.yaml.
double _d(Object? v) => (v as num?)?.toDouble() ?? 0;
int _i(Object? v) => (v as num?)?.toInt() ?? 0;
Map<String, Object?> _m(Object? v) => (v as Map?)?.cast<String, Object?>() ?? const {};
List<Map<String, Object?>> _list(Object? v) => ((v as List?) ?? const []).map(_m).toList();
DateTime _t(Object? v) => DateTime.tryParse(v as String? ?? '') ?? DateTime.fromMillisecondsSinceEpoch(0);

class UserSettings {
  const UserSettings({
    this.language = 'th',
    this.currency = 'THB',
    this.theme = 'light',
    this.push = true,
    this.email = true,
    this.trip = true,
    this.promo = true,
  });
  final String language, currency, theme;
  final bool push, email, trip, promo;

  UserSettings copyWith({String? language, String? currency, String? theme, bool? push, bool? email, bool? trip, bool? promo}) => UserSettings(
        language: language ?? this.language,
        currency: currency ?? this.currency,
        theme: theme ?? this.theme,
        push: push ?? this.push,
        email: email ?? this.email,
        trip: trip ?? this.trip,
        promo: promo ?? this.promo,
      );

  factory UserSettings.fromJson(Map<String, Object?> j) {
    final n = _m(j['notifications']);
    return UserSettings(
      language: j['language'] as String? ?? 'th',
      currency: j['currency'] as String? ?? 'THB',
      theme: j['theme'] as String? ?? 'light',
      push: n['push'] as bool? ?? true,
      email: n['email'] as bool? ?? true,
      trip: n['trip'] as bool? ?? true,
      promo: n['promo'] as bool? ?? true,
    );
  }

  Map<String, Object?> toJson() => {
        'language': language,
        'currency': currency,
        'theme': theme,
        'notifications': {'push': push, 'email': email, 'trip': trip, 'promo': promo},
      };
}

class Coupon {
  const Coupon({required this.id, required this.code, required this.title, required this.detail, required this.minSpend, required this.maxDiscount, required this.expiry, required this.status});
  final int id;
  final String code, title, detail, expiry, status;
  final double minSpend, maxDiscount;
  bool get usable => status == 'available';
  factory Coupon.fromJson(Map<String, Object?> j) => Coupon(
        id: _i(j['id']),
        code: j['code'] as String? ?? '',
        title: j['title'] as String? ?? '',
        detail: j['detail'] as String? ?? '',
        minSpend: _d(j['minSpend']),
        maxDiscount: _d(j['maxDiscount']),
        expiry: j['expiry'] as String? ?? '',
        status: j['status'] as String? ?? 'available',
      );
}

class AppNotification {
  const AppNotification({required this.id, required this.type, required this.title, required this.message, required this.read, required this.createdAt});
  final int id;
  final String type, title, message;
  final bool read;
  final DateTime createdAt;
  AppNotification asRead() => AppNotification(id: id, type: type, title: title, message: message, read: true, createdAt: createdAt);
  factory AppNotification.fromJson(Map<String, Object?> j) => AppNotification(
        id: _i(j['id']),
        type: j['type'] as String? ?? 'system',
        title: j['title'] as String? ?? '',
        message: j['message'] as String? ?? '',
        read: j['read'] as bool? ?? false,
        createdAt: _t(j['createdAt']),
      );
}

class ItineraryItem {
  const ItineraryItem({required this.time, required this.text, required this.place});
  final String time, text, place;
  factory ItineraryItem.fromJson(Map<String, Object?> j) =>
      ItineraryItem(time: j['time'] as String? ?? '', text: j['text'] as String? ?? '', place: j['place'] as String? ?? '');
  Map<String, Object?> toJson() => {'time': time, 'text': text, 'place': place};
}

class ItineraryDay {
  const ItineraryDay({required this.title, required this.items});
  final String title;
  final List<ItineraryItem> items;
  factory ItineraryDay.fromJson(Map<String, Object?> j) =>
      ItineraryDay(title: j['title'] as String? ?? '', items: _list(j['items']).map(ItineraryItem.fromJson).toList());
  Map<String, Object?> toJson() => {'title': title, 'items': items.map((e) => e.toJson()).toList()};
}

class BudgetLine {
  const BudgetLine(this.label, this.value);
  final String label;
  final double value;
  factory BudgetLine.fromJson(Map<String, Object?> j) => BudgetLine(j['label'] as String? ?? '', _d(j['value']));
  Map<String, Object?> toJson() => {'label': label, 'value': value};
}

class ItineraryResult {
  const ItineraryResult({required this.days, required this.budget});
  final List<ItineraryDay> days;
  final List<BudgetLine> budget;
  factory ItineraryResult.fromJson(Map<String, Object?> j) => ItineraryResult(
        days: _list(j['days']).map(ItineraryDay.fromJson).toList(),
        budget: _list(j['budgetBreakdown']).map(BudgetLine.fromJson).toList(),
      );
  Map<String, Object?> toJson() => {'days': days.map((e) => e.toJson()).toList(), 'budgetBreakdown': budget.map((e) => e.toJson()).toList()};
}

class ItineraryRequest {
  const ItineraryRequest({required this.destination, required this.days, required this.budget, this.styles = const [], this.travelers = 1});
  final String destination;
  final int days, travelers;
  final double budget;
  final List<String> styles;
  Map<String, Object?> toJson() => {'destination': destination, 'days': days, 'budget': budget, 'styles': styles, 'travelers': travelers};
}

class ItineraryJob {
  const ItineraryJob({required this.jobId, required this.status, required this.degraded, this.result});
  final String jobId, status;
  final bool degraded;
  final ItineraryResult? result;
  factory ItineraryJob.fromJson(Map<String, Object?> j) => ItineraryJob(
        jobId: j['jobId'] as String? ?? '',
        status: j['status'] as String? ?? 'pending',
        degraded: j['degraded'] as bool? ?? false,
        result: j['result'] == null ? null : ItineraryResult.fromJson(_m(j['result'])),
      );
}

class SavedTrip {
  const SavedTrip({required this.id, required this.destination, required this.days, required this.travelers, required this.budget, required this.source, this.itinerary});
  final int id, days, travelers;
  final String destination, source;
  final double budget;
  final ItineraryResult? itinerary;
  factory SavedTrip.fromJson(Map<String, Object?> j) => SavedTrip(
        id: _i(j['id']),
        destination: j['destination'] as String? ?? '',
        days: _i(j['days']),
        travelers: _i(j['travelers']),
        budget: _d(j['budget']),
        source: j['source'] as String? ?? 'manual',
        itinerary: j['itinerary'] == null ? null : ItineraryResult.fromJson(_m(j['itinerary'])),
      );
}

class Review {
  const Review({required this.id, required this.itemName, required this.itemType, required this.rating, required this.comment, required this.createdAt});
  final int id, rating;
  final String itemName, itemType, comment;
  final DateTime createdAt;
  factory Review.fromJson(Map<String, Object?> j) => Review(
        id: _i(j['id']),
        itemName: j['itemName'] as String? ?? '',
        itemType: j['itemType'] as String? ?? '',
        rating: _i(j['rating']),
        comment: j['comment'] as String? ?? '',
        createdAt: _t(j['createdAt']),
      );
}

class SupportTicket {
  const SupportTicket({required this.id, required this.subject, required this.message, required this.createdAt});
  final int id;
  final String subject, message;
  final DateTime createdAt;
  factory SupportTicket.fromJson(Map<String, Object?> j) => SupportTicket(
        id: _i(j['id']),
        subject: j['subject'] as String? ?? '',
        message: j['message'] as String? ?? '',
        createdAt: _t(j['createdAt']),
      );
}

class Post {
  const Post({
    required this.id,
    required this.authorId,
    required this.type,
    required this.destination,
    required this.text,
    required this.likeCount,
    required this.hasImage,
    required this.createdAt,
    this.liked = false,
    this.saved = false,
    this.sharedDays,
    this.authorName = '',
    this.commentCount = 0,
    this.groupId,
  });
  final int id, likeCount, commentCount;
  final String authorId, authorName, type, destination, text;
  final bool hasImage, liked, saved;
  final int? sharedDays, groupId;
  final DateTime createdAt;

  /// Name to show: the server's display name, else the member id without its `m_` prefix.
  String get displayName => authorName.isNotEmpty ? authorName : authorId.replaceFirst('m_', '');

  Post copyWith({bool? liked, int? likeCount, bool? saved, int? commentCount}) => Post(
      id: id,
      authorId: authorId,
      authorName: authorName,
      type: type,
      destination: destination,
      text: text,
      likeCount: likeCount ?? this.likeCount,
      commentCount: commentCount ?? this.commentCount,
      hasImage: hasImage,
      createdAt: createdAt,
      liked: liked ?? this.liked,
      saved: saved ?? this.saved,
      sharedDays: sharedDays,
      groupId: groupId);
  Post withLike(bool liked, int count) => copyWith(liked: liked, likeCount: count);
  factory Post.fromJson(Map<String, Object?> j) => Post(
        id: _i(j['id']),
        authorId: j['authorId'] as String? ?? '',
        authorName: j['authorName'] as String? ?? '',
        type: j['type'] as String? ?? 'story',
        destination: j['destination'] as String? ?? '',
        text: j['text'] as String? ?? '',
        likeCount: _i(j['likeCount']),
        commentCount: _i(j['commentCount']),
        hasImage: j['hasImage'] as bool? ?? false,
        liked: j['liked'] as bool? ?? false,
        saved: j['saved'] as bool? ?? false,
        createdAt: _t(j['createdAt']),
        groupId: j['groupId'] == null ? null : _i(j['groupId']),
        sharedDays: j['sharedTrip'] == null ? null : _i(_m(j['sharedTrip'])['days']),
      );
}

class Comment {
  const Comment({required this.id, required this.authorId, required this.authorName, required this.text, required this.createdAt});
  final int id;
  final String authorId, authorName, text;
  final DateTime createdAt;
  String get displayName => authorName.isNotEmpty ? authorName : authorId.replaceFirst('m_', '');
  factory Comment.fromJson(Map<String, Object?> j) => Comment(
        id: _i(j['id']),
        authorId: j['authorId'] as String? ?? '',
        authorName: j['authorName'] as String? ?? '',
        text: j['text'] as String? ?? '',
        createdAt: _t(j['createdAt']),
      );
}

class PostDetail {
  const PostDetail(this.post, this.comments);
  final Post post;
  final List<Comment> comments;
  factory PostDetail.fromJson(Map<String, Object?> j) => PostDetail(Post.fromJson(j), _list(j['comments']).map(Comment.fromJson).toList());
}

class Group {
  const Group({required this.id, required this.name, required this.description, this.destination, this.memberCount = 0, this.joined = false, this.rules = const []});
  final int id, memberCount;
  final String name, description;
  final String? destination;
  final bool joined;
  final List<String> rules;
  Group copyWith({bool? joined, int? memberCount}) => Group(
      id: id, name: name, description: description, destination: destination, memberCount: memberCount ?? this.memberCount, joined: joined ?? this.joined, rules: rules);
  factory Group.fromJson(Map<String, Object?> j) => Group(
        id: _i(j['id']),
        name: j['name'] as String? ?? '',
        description: j['description'] as String? ?? '',
        destination: j['destination'] as String?,
        memberCount: _i(j['memberCount']),
        joined: j['joined'] as bool? ?? false,
        rules: ((j['rules'] as List?) ?? const []).map((e) => '$e').toList(),
      );
}

class LikeResult {
  const LikeResult(this.liked, this.likeCount);
  final bool liked;
  final int likeCount;
}
