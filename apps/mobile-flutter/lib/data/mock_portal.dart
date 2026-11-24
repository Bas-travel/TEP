import 'api_client.dart';
import 'mock_fixtures.dart';
import 'models.dart';
import 'models_engage.dart';

/// In-memory back-office for API_MODE=mock: mirrors the rules of services/api-go (handlers_partner/admin/staff.go)
/// so the portals can be developed and demoed without a backend. The server remains the source of truth for RBAC.
/// Limitation: the traveler catalog of the mock client is static, so listing edits are visible in the portals only.
class MockPortal {
  MockPortal(this._now) {
    final now = _now();
    for (final l in kMockListings.take(12)) {
      _listings.add({
        'id': l['id'],
        'name': l['name'],
        'type': l['type'],
        'location': l['location'],
        'destinationId': l['destinationId'],
        'priceFrom': l['priceFrom'] ?? 1000,
        'description': l['description'] ?? '',
        'amenities': List<Object?>.of((l['amenities'] as List?) ?? const []),
        'mapQuery': l['mapQuery'] ?? '',
        'imageUrl': '',
        'status': 'published',
        'partnerId': (l['id'] as int) <= 4 ? 1 : 0,
      });
    }
    _users.addAll([
      _user(1, 'demo@tep.dev', 'สวัสดี อนันต์', 'traveler', now),
      _user(2, 'm_nok@members.tep.invalid', 'นก ใจดี', 'traveler', now),
      _user(3, 'm_ton@members.tep.invalid', 'ต้น สายเที่ยว', 'traveler', now),
      _user(4, 'm_fah@members.tep.invalid', 'ฟ้า ใสใจ', 'traveler', now),
      _user(5, 'm_pim@members.tep.invalid', 'พิม กินเก่ง', 'traveler', now),
      _user(6, 'admin@tep.dev', 'แอดมิน ระบบ', 'admin', now),
      _user(7, 'staff@tep.dev', 'สมชาย ซัพพอร์ต', 'staff', now),
      _user(8, 'partner@tep.dev', 'วิภา อันดามัน', 'partner', now),
    ]);
    _partners.addAll([
      {'id': 1, 'userId': 8, 'businessName': 'Andaman Partner Co.', 'contactPhone': '076-000-000', 'description': 'โรงแรมและกิจกรรมฝั่งอันดามัน', 'status': 'approved', 'commissionRate': 12.0, 'decisionNote': '', 'createdAt': now.subtract(const Duration(days: 30)).toIso8601String()},
      {'id': 2, 'userId': 4, 'businessName': 'Fah Local Tours', 'contactPhone': '081-000-0000', 'description': 'ทัวร์ท้องถิ่นกระบี่', 'status': 'pending', 'commissionRate': 12.0, 'decisionNote': '', 'createdAt': now.subtract(const Duration(hours: 6)).toIso8601String()},
    ]);
    _bookings.addAll([
      _booking(1, 2, 1, 'confirmed', now.subtract(const Duration(hours: 3))),
      _booking(2, 3, 2, 'completed', now.subtract(const Duration(hours: 30))),
      _booking(3, 2, 3, 'pending_payment', now.subtract(const Duration(hours: 1))),
    ]);
    (_bookings[1]['checks'] as List).add({'id': 1, 'staffName': 'สมชาย ซ.', 'result': 'verified', 'note': 'ตรวจสอบการชำระเงินแล้ว', 'createdAt': now.subtract(const Duration(hours: 26)).toIso8601String()});
    _nextCheck = 2;
    _tickets.addAll([
      _ticket(1, 5, 'สอบถามการยกเลิกการจอง', 'ต้องการยกเลิกการจองโรงแรมที่ภูเก็ต ทำอย่างไรคะ', 'open', now.subtract(const Duration(hours: 2))),
      _ticket(2, 2, 'คูปองใช้ไม่ได้', 'ใส่โค้ด WELCOME10 แล้วระบบแจ้งว่าไม่ผ่าน', 'answered', now.subtract(const Duration(hours: 30))),
      _ticket(3, 3, 'ขอใบกำกับภาษี', 'ขอใบกำกับภาษีสำหรับการจองที่ผ่านมา', 'resolved', now.subtract(const Duration(hours: 72))),
    ]);
    (_tickets[1]['replies'] as List).add(_reply(1, 'staff', 'สมชาย ซ.', 'ทีมงานได้รับเรื่องแล้ว กำลังตรวจสอบให้ค่ะ', now.subtract(const Duration(hours: 29))));
    (_tickets[2]['replies'] as List).add(_reply(2, 'staff', 'สมชาย ซ.', 'ส่งใบกำกับภาษีให้ทางอีเมลแล้วค่ะ', now.subtract(const Duration(hours: 71))));
    _nextReply = 3;
    _nextTicket = 4;
  }

  final DateTime Function() _now;
  final List<Json> _listings = [];
  final List<Json> _users = [];
  final List<Json> _partners = [];
  final List<Json> _bookings = [];
  final List<Json> _tickets = [];
  final List<Json> _audits = [];
  int _nextListing = 13, _nextCheck = 1, _nextReply = 1, _nextTicket = 1, _nextAudit = 1, _nextPartner = 3;

  // ---------- seeds ----------
  Json _user(int id, String email, String name, String role, DateTime now) =>
      {'id': id, 'handle': 'm_$id', 'email': email, 'name': name, 'role': role, 'status': 'active', 'createdAt': now.subtract(Duration(days: 40 - id)).toIso8601String()};

  Json _booking(int id, int userId, int listingId, String status, DateTime date) {
    final l = _listings.firstWhere((e) => e['id'] == listingId);
    final price = (l['priceFrom'] as num).toDouble();
    final fee = _r2(price * 0.03);
    final u = _users.isEmpty ? null : _users.where((e) => e['id'] == userId).firstOrNull;
    return {
      'id': id,
      'bookingNumber': 'TEP-${date.year % 100}${date.month.toString().padLeft(2, '0')}${date.day.toString().padLeft(2, '0')}-${id.toString().padLeft(5, '0')}',
      'type': l['type'],
      'title': l['name'],
      'subtitle': l['location'],
      'status': status,
      'date': date.toIso8601String(),
      'guests': 2,
      'userId': userId,
      'travelerName': u?['name'] ?? '',
      'travelerEmail': u?['email'] ?? '',
      'subtotal': price,
      'serviceFee': fee,
      'amount': _r2(price + fee),
      'cancelReason': '',
      'lines': [
        {'listingId': listingId, 'quantity': 1, 'unitPrice': price}
      ],
      'checks': <Object?>[],
    };
  }

  Json _ticket(int id, int userId, String subject, String message, String status, DateTime at) =>
      {'id': id, 'userId': userId, 'subject': subject, 'message': message, 'status': status, 'createdAt': at.toIso8601String(), 'replies': <Object?>[], 'firstRepliedAt': null};

  Json _reply(int id, String role, String name, String message, DateTime at) =>
      {'id': id, 'authorRole': role, 'authorName': name, 'message': message, 'createdAt': at.toIso8601String()};

  // ---------- helpers ----------
  double _r2(double x) => (x * 100).roundToDouble() / 100;

  ApiException _err(int status, String code, String title, [String? detail]) => ApiException(status: status, code: code, title: title, detail: detail);
  ApiException _forbidden(String d) => _err(403, 'AUTH_FORBIDDEN', 'Forbidden', d);
  ApiException _notFound() => _err(404, 'NOT_FOUND', 'Not Found', 'resource not found');
  ApiException _conflict(String d) => _err(409, 'CONFLICT', 'Conflict', d);
  ApiException _inv(String field, String msg) =>
      ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: [FieldError(field, msg)]);

  void _need(UserProfile u, List<String> roles) {
    final role = _users.where((e) => e['id'] == u.id).firstOrNull?['role'] as String? ?? u.role;
    if (!roles.contains(role)) throw _forbidden('this area requires role ${roles.join(' or ')}');
  }

  Json? _partnerOf(int userId) => _partners.where((p) => p['userId'] == userId).firstOrNull;

  Json _myPartner(UserProfile u) {
    _need(u, ['partner']);
    final p = _partnerOf(u.id);
    if (p == null) throw _forbidden('no partner profile');
    if (p['status'] != 'approved') throw _forbidden('partner account is ${p['status']}');
    return p;
  }

  Json _find(List<Json> l, int id) {
    final r = l.where((e) => e['id'] == id).firstOrNull;
    if (r == null) throw _notFound();
    return r;
  }

  Json _pageOf(List<Json> all, Map<String, String?> q) {
    final limit = int.tryParse(q['limit'] ?? '') ?? 20;
    final offset = int.tryParse(q['offset'] ?? '') ?? 0;
    if (limit < 1 || limit > 100) throw _inv('limit', 'must be 1..100');
    if (offset < 0) throw _inv('offset', 'must be >= 0');
    final end = (offset + limit) > all.length ? all.length : offset + limit;
    final items = offset >= all.length ? <Json>[] : all.sublist(offset, end);
    return {'items': items, 'total': all.length, 'offset': offset, 'limit': limit};
  }

  bool _has(Json m, String field, String? q) => q == null || q.isEmpty || '${m[field]}'.toLowerCase().contains(q.toLowerCase());

  void _audit(UserProfile u, String action, String entity, int id, [Json? meta]) {
    final name = _users.where((e) => e['id'] == u.id).firstOrNull?['name'] ?? u.fullName;
    _audits.insert(0, {'id': _nextAudit++, 'actorId': u.id, 'actorName': name, 'action': action, 'entity': entity, 'entityId': '$id', 'meta': meta ?? <String, Object?>{}, 'createdAt': _now().toIso8601String()});
  }

  Set<int> _ownIds(int partnerId) => {for (final l in _listings) if (l['partnerId'] == partnerId) l['id'] as int};

  double _share(Json b, Set<int> own) {
    var g = 0.0;
    for (final ln in (b['lines'] as List).cast<Json>()) {
      if (own.contains(ln['listingId'])) g += (ln['quantity'] as int) * (ln['unitPrice'] as num).toDouble();
    }
    return _r2(g);
  }

  bool _allMine(Json b, Set<int> own) {
    final lines = (b['lines'] as List).cast<Json>();
    return lines.isNotEmpty && lines.every((ln) => own.contains(ln['listingId']));
  }

  bool _touches(Json b, Set<int> own) => (b['lines'] as List).cast<Json>().any((ln) => own.contains(ln['listingId']));

  // ---------- traveler side ----------
  List<SupportTicket> ticketsOf(int userId) =>
      _tickets.where((t) => t['userId'] == userId).map((t) => SupportTicket.fromJson(_ticketOut(t))).toList().reversed.toList();

  SupportTicket addTicket(UserProfile u, String subject, String message) {
    final t = _ticket(_nextTicket++, u.id, subject, message, 'open', _now());
    _tickets.add(t);
    return SupportTicket.fromJson(_ticketOut(t));
  }

  Json _ticketOut(Json t) => {
        'id': t['id'],
        'subject': t['subject'],
        'message': t['message'],
        'status': t['status'],
        'createdAt': t['createdAt'],
        'replies': List<Object?>.of(t['replies'] as List),
      };

  // ---------- dispatcher ----------
  Json call(String method, String path, {required UserProfile user, Map<String, String?> query = const {}, Object? body}) {
    final ids = <int>[];
    final norm = path.replaceAllMapped(RegExp(r'/(\d+)'), (m) {
      ids.add(int.parse(m[1]!));
      return '/{}';
    });
    final b = body is Map ? body.cast<String, Object?>() : <String, Object?>{};
    switch ('$method $norm') {
      // traveler
      case 'POST /v1/partners/apply':
        return _apply(user, b);
      case 'GET /v1/partners/me':
        return _partnerView(_partnerOf(user.id) ?? (throw _notFound()), false);
      case 'POST /v1/support/tickets/{}/replies':
        return _travelerReply(user, ids[0], b);
      // partner
      case 'GET /partner/v1/dashboard':
        return _pDashboard(_myPartner(user));
      case 'GET /partner/v1/profile':
        return _partnerView(_myPartner(user), false);
      case 'PATCH /partner/v1/profile':
        return _pProfile(user, _myPartner(user), b);
      case 'GET /partner/v1/listings':
        return _pListings(_myPartner(user), query);
      case 'POST /partner/v1/listings':
        return _pSaveListing(user, _myPartner(user), null, b);
      case 'GET /partner/v1/listings/{}':
        return _ownListing(_myPartner(user), ids[0]);
      case 'PATCH /partner/v1/listings/{}':
        return _pSaveListing(user, _myPartner(user), ids[0], b);
      case 'GET /partner/v1/bookings':
        return _pBookings(_myPartner(user), query);
      case 'GET /partner/v1/bookings/{}':
        return _pBookingGet(_myPartner(user), ids[0]);
      case 'POST /partner/v1/bookings/{}/complete':
        return _pAct(user, _myPartner(user), ids[0], 'complete', b);
      case 'POST /partner/v1/bookings/{}/reject':
        return _pAct(user, _myPartner(user), ids[0], 'reject', b);
      case 'GET /partner/v1/earnings':
        return _earnings(_myPartner(user));
      // admin
      case 'GET /admin/v1/dashboard':
        _need(user, ['admin']);
        return _aDashboard();
      case 'GET /admin/v1/users':
        _need(user, ['admin']);
        return _aUsers(query);
      case 'GET /admin/v1/users/{}':
        _need(user, ['admin']);
        return _find(_users, ids[0]);
      case 'PATCH /admin/v1/users/{}':
        _need(user, ['admin']);
        return _aPatchUser(user, ids[0], b);
      case 'GET /admin/v1/partners':
        _need(user, ['admin']);
        return _aPartners(query);
      case 'GET /admin/v1/partners/{}':
        _need(user, ['admin']);
        return _partnerView(_find(_partners, ids[0]), true);
      case 'PATCH /admin/v1/partners/{}':
        _need(user, ['admin']);
        return _aPatchPartner(user, ids[0], b);
      case 'GET /admin/v1/listings':
        _need(user, ['admin']);
        return _aListings(query);
      case 'PATCH /admin/v1/listings/{}':
        _need(user, ['admin']);
        return _aPatchListing(user, ids[0], b);
      case 'GET /admin/v1/audit-logs':
        _need(user, ['admin']);
        return _aAudit(query);
      case 'GET /admin/v1/system':
        _need(user, ['admin']);
        return {'version': 'mock', 'env': 'dev', 'mode': 'mock', 'serverTime': _now().toUtc().toIso8601String(), 'dependencies': {'mysql': 'disabled', 'redis': 'disabled'}};
      // staff
      case 'GET /staff/v1/dashboard':
        _need(user, ['staff', 'admin']);
        return _sCounters(_startOfDay(_now()));
      case 'GET /staff/v1/bookings':
        _need(user, ['staff', 'admin']);
        return _sBookings(query);
      case 'GET /staff/v1/bookings/{}':
        _need(user, ['staff', 'admin']);
        return _staffBooking(_find(_bookings, ids[0]));
      case 'POST /staff/v1/bookings/{}/check':
        _need(user, ['staff', 'admin']);
        return _sCheck(user, ids[0], b);
      case 'GET /staff/v1/tickets':
        _need(user, ['staff', 'admin']);
        return _sTickets(query);
      case 'GET /staff/v1/tickets/{}':
        _need(user, ['staff', 'admin']);
        return _staffTicket(_find(_tickets, ids[0]));
      case 'PATCH /staff/v1/tickets/{}':
        _need(user, ['staff', 'admin']);
        return _sPatchTicket(user, ids[0], b);
      case 'POST /staff/v1/tickets/{}/replies':
        _need(user, ['staff', 'admin']);
        return _sReply(user, ids[0], b);
      case 'GET /staff/v1/reports/daily':
        _need(user, ['staff', 'admin']);
        return _sReport(user, query);
    }
    throw _notFound();
  }

  // ---------- partners ----------
  Json _partnerView(Json p, bool owner) {
    final u = _users.where((e) => e['id'] == p['userId']).firstOrNull;
    return {
      ...p,
      'listingCount': _listings.where((l) => l['partnerId'] == p['id']).length,
      if (owner) 'ownerName': u?['name'] ?? '',
      if (owner) 'ownerEmail': u?['email'] ?? '',
    };
  }

  Json _apply(UserProfile u, Json b) {
    final name = '${b['businessName'] ?? ''}'.trim();
    if (name.isEmpty || name.length > 120) throw _inv('businessName', 'required, max 120 characters');
    final role = _users.where((e) => e['id'] == u.id).firstOrNull?['role'] ?? u.role;
    if (role != 'traveler') throw _conflict('only traveler accounts can apply to become a partner');
    final old = _partnerOf(u.id);
    if (old != null) {
      if (old['status'] != 'rejected') throw _conflict('a partner application already exists (${old['status']})');
      old
        ..['businessName'] = name
        ..['contactPhone'] = '${b['contactPhone'] ?? ''}'.trim()
        ..['description'] = '${b['description'] ?? ''}'.trim()
        ..['status'] = 'pending'
        ..['decisionNote'] = '';
      return _partnerView(old, false);
    }
    final p = {
      'id': _nextPartner++,
      'userId': u.id,
      'businessName': name,
      'contactPhone': '${b['contactPhone'] ?? ''}'.trim(),
      'description': '${b['description'] ?? ''}'.trim(),
      'status': 'pending',
      'commissionRate': 12.0,
      'decisionNote': '',
      'createdAt': _now().toIso8601String(),
    };
    _partners.add(p);
    return _partnerView(p, false);
  }

  Json _pProfile(UserProfile u, Json p, Json b) {
    final name = (b['businessName'] ?? p['businessName']).toString().trim();
    if (name.isEmpty || name.length > 120) throw _inv('businessName', 'required, max 120 characters');
    p
      ..['businessName'] = name
      ..['contactPhone'] = (b['contactPhone'] ?? p['contactPhone']).toString().trim()
      ..['description'] = (b['description'] ?? p['description']).toString().trim();
    _audit(u, 'partner.profile.update', 'partner', p['id'] as int);
    return _partnerView(p, false);
  }

  // ---------- partner listings ----------
  Json _listingOut(Json l) => {...l, 'amenities': List<Object?>.of(l['amenities'] as List)};

  Json _pListings(Json p, Map<String, String?> q) {
    final st = q['status'];
    if (st != null && st.isNotEmpty && !const ['draft', 'published', 'archived'].contains(st)) throw _inv('status', 'invalid status');
    final items = _listings
        .where((l) => l['partnerId'] == p['id'] && (st == null || st.isEmpty || l['status'] == st) && (q['type'] == null || q['type']!.isEmpty || l['type'] == q['type']) && (_has(l, 'name', q['q']) || _has(l, 'location', q['q'])))
        .map(_listingOut)
        .toList();
    return {'items': items};
  }

  Json _ownListing(Json p, int id) {
    final l = _find(_listings, id);
    if (l['partnerId'] != p['id']) throw _notFound();
    return _listingOut(l);
  }

  Json _pSaveListing(UserProfile u, Json p, int? id, Json b) {
    final Json l = id == null
        ? {'id': 0, 'name': '', 'type': '', 'location': '', 'destinationId': 0, 'priceFrom': 0, 'description': '', 'amenities': <Object?>[], 'mapQuery': '', 'imageUrl': '', 'status': 'draft', 'partnerId': p['id']}
        : Map<String, Object?>.of(_ownListing(p, id));
    for (final k in const ['name', 'type', 'location', 'destinationId', 'priceFrom', 'description', 'amenities', 'mapQuery', 'imageUrl', 'status']) {
      if (b.containsKey(k) && b[k] != null) l[k] = b[k];
    }
    final errs = <FieldError>[];
    final name = '${l['name']}'.trim();
    if (name.isEmpty || name.length > 190) errs.add(const FieldError('name', 'required, max 190 characters'));
    if (!const ['hotel', 'restaurant', 'attraction'].contains(l['type'])) errs.add(const FieldError('type', 'must be hotel, restaurant or attraction'));
    if ('${l['location']}'.trim().isEmpty) errs.add(const FieldError('location', 'required'));
    final price = (l['priceFrom'] as num).toDouble();
    if (price <= 0 || price > 1000000) errs.add(const FieldError('priceFrom', 'must be > 0 and <= 1,000,000'));
    if ('${l['description']}'.length > 2000) errs.add(const FieldError('description', 'max 2000 characters'));
    if (!const ['draft', 'published', 'archived'].contains(l['status'])) errs.add(const FieldError('status', 'must be draft, published or archived'));
    if (!kMockDestinations.any((d) => d['id'] == l['destinationId'])) errs.add(const FieldError('destinationId', 'unknown destination'));
    if (errs.isNotEmpty) throw ApiException(status: 422, code: 'VALIDATION_FAILED', title: 'Validation failed', errors: errs);
    l['name'] = name;
    l['priceFrom'] = _r2(price);
    if (id == null) {
      l['id'] = _nextListing++;
      _listings.add(l);
    } else {
      _listings[_listings.indexWhere((e) => e['id'] == id)] = l;
    }
    _audit(u, id == null ? 'partner.listing.create' : 'partner.listing.update', 'listing', l['id'] as int, {'status': l['status']});
    return _listingOut(l);
  }

  // ---------- partner bookings & earnings ----------
  Json _pBookingView(Json p, Set<int> own, Json b) {
    final gross = _share(b, own);
    final rate = (p['commissionRate'] as num).toDouble();
    final comm = _r2(gross * rate / 100);
    final open = b['status'] == 'confirmed' || b['status'] == 'pending_payment';
    return {
      'id': b['id'],
      'bookingNumber': b['bookingNumber'],
      'type': b['type'],
      'title': b['title'],
      'status': b['status'],
      'date': b['date'],
      'guests': b['guests'],
      'travelerName': b['travelerName'],
      'gross': gross,
      'commission': comm,
      'net': _r2(gross - comm),
      'canAct': _allMine(b, own) && open,
      'cancelReason': b['cancelReason'],
    };
  }

  Json _pBookings(Json p, Map<String, String?> q) {
    final own = _ownIds(p['id'] as int);
    final st = q['status'];
    if (st != null && st.isNotEmpty && !const ['pending_payment', 'confirmed', 'completed', 'cancelled'].contains(st)) throw _inv('status', 'invalid status');
    final list = _bookings
        .where((b) => _touches(b, own) && (st == null || st.isEmpty || b['status'] == st) && (_has(b, 'bookingNumber', q['q']) || _has(b, 'title', q['q'])))
        .toList()
      ..sort((a, c) => '${c['date']}'.compareTo('${a['date']}'));
    final page = _pageOf(list, q);
    return {...page, 'items': (page['items'] as List).cast<Json>().map((b) => _pBookingView(p, own, b)).toList()};
  }

  Json _pBookingGet(Json p, int id) {
    final own = _ownIds(p['id'] as int);
    final b = _find(_bookings, id);
    if (!_touches(b, own)) throw _notFound();
    return _pBookingView(p, own, b);
  }

  Json _pAct(UserProfile u, Json p, int id, String action, Json body) {
    final own = _ownIds(p['id'] as int);
    final b = _find(_bookings, id);
    if (!_touches(b, own)) throw _notFound();
    final reason = '${body['reason'] ?? ''}'.trim();
    if (action == 'reject' && (reason.isEmpty || reason.length > 450)) throw _inv('reason', 'required, max 450 characters');
    if (!_allMine(b, own)) throw _conflict('this booking includes items of other providers; contact support to change it');
    if (action == 'complete') {
      if (b['status'] != 'confirmed') throw _conflict('only a confirmed booking can be completed (now ${b['status']})');
      b['status'] = 'completed';
    } else {
      if (b['status'] != 'pending_payment' && b['status'] != 'confirmed') throw _conflict('booking cannot be rejected in state ${b['status']}');
      b['status'] = 'cancelled';
      b['cancelReason'] = 'ผู้ให้บริการปฏิเสธ: $reason';
    }
    _audit(u, 'partner.booking.$action', 'booking', id);
    return _pBookingView(p, own, b);
  }

  Json _earnings(Json p) {
    final own = _ownIds(p['id'] as int);
    final rate = (p['commissionRate'] as num).toDouble();
    final byMonth = <String, List<double>>{};
    final list = _bookings.where((b) => _touches(b, own) && (b['status'] == 'confirmed' || b['status'] == 'completed')).toList()
      ..sort((a, c) => '${c['date']}'.compareTo('${a['date']}'));
    for (final b in list) {
      final m = '${b['date']}'.substring(0, 7);
      final e = byMonth.putIfAbsent(m, () => [0, 0]);
      e[0] += 1;
      e[1] += _share(b, own);
    }
    var g = 0.0, c = 0.0, n = 0.0;
    final months = <Json>[];
    byMonth.forEach((m, e) {
      final gross = _r2(e[1]);
      final comm = _r2(gross * rate / 100);
      months.add({'month': m, 'bookings': e[0].toInt(), 'gross': gross, 'commission': comm, 'net': _r2(gross - comm)});
      g += gross;
      c += comm;
      n += gross - comm;
    });
    return {'commissionRate': rate, 'gross': _r2(g), 'commission': _r2(c), 'net': _r2(n), 'months': months};
  }

  Json _pDashboard(Json p) {
    final own = _ownIds(p['id'] as int);
    final by = {'draft': 0, 'published': 0, 'archived': 0};
    for (final l in _listings.where((l) => l['partnerId'] == p['id'])) {
      by['${l['status']}'] = (by['${l['status']}'] ?? 0) + 1;
    }
    int count(String st) => _bookings.where((b) => _touches(b, own) && b['status'] == st).length;
    final e = _earnings(p);
    final cur = _now().toIso8601String().substring(0, 7);
    final month = (e['months'] as List).cast<Json>().where((m) => m['month'] == cur).firstOrNull ?? {'month': cur, 'bookings': 0, 'gross': 0.0, 'commission': 0.0, 'net': 0.0};
    return {
      'businessName': p['businessName'],
      'commissionRate': p['commissionRate'],
      'listings': by,
      'bookingsToComplete': count('confirmed'),
      'bookingsAwaitingPayment': count('pending_payment'),
      'thisMonth': month,
      'totalNet': e['net'],
    };
  }

  // ---------- admin ----------
  DateTime _startOfDay(DateTime t) => DateTime(t.year, t.month, t.day);

  Json _aDashboard() {
    final byRole = <String, int>{};
    for (final u in _users.where((u) => u['status'] == 'active')) {
      byRole['${u['role']}'] = (byRole['${u['role']}'] ?? 0) + 1;
    }
    final partners = <String, int>{};
    for (final p in _partners) {
      partners['${p['status']}'] = (partners['${p['status']}'] ?? 0) + 1;
    }
    final since = _startOfDay(_now());
    var revenue = 0.0;
    for (final b in _bookings.where((b) => b['status'] == 'confirmed' || b['status'] == 'completed')) {
      revenue += (b['amount'] as num).toDouble();
    }
    return {
      'users': _users.length,
      'usersSuspended': _users.where((u) => u['status'] == 'suspended').length,
      'usersByRole': byRole,
      'partners': partners,
      'listingsPublished': _listings.where((l) => l['status'] == 'published').length,
      'bookingsToday': _bookings.where((b) => !DateTime.parse('${b['date']}').isBefore(since)).length,
      'bookingsTotal': _bookings.length,
      'revenue': _r2(revenue),
    };
  }

  Json _aUsers(Map<String, String?> q) {
    final role = q['role'], st = q['status'];
    if (role != null && role.isNotEmpty && !const ['traveler', 'partner', 'staff', 'moderator', 'admin'].contains(role)) throw _inv('role', 'invalid role');
    if (st != null && st.isNotEmpty && !const ['active', 'suspended', 'deleted'].contains(st)) throw _inv('status', 'invalid status');
    final list = _users
        .where((u) => (role == null || role.isEmpty || u['role'] == role) && (st == null || st.isEmpty || u['status'] == st) && (_has(u, 'name', q['q']) || _has(u, 'email', q['q'])))
        .toList();
    return _pageOf(list, q);
  }

  Json _aPatchUser(UserProfile admin, int id, Json b) {
    final role = b['role'], status = b['status'];
    if (role == null && status == null) throw _inv('body', 'role or status required');
    if (role != null && !const ['traveler', 'staff', 'moderator', 'admin'].contains(role)) {
      throw _inv('role', 'must be traveler, staff, moderator or admin (partners are managed under /admin/v1/partners)');
    }
    if (status != null && !const ['active', 'suspended'].contains(status)) throw _inv('status', 'must be active or suspended');
    final u = _find(_users, id);
    if (u['status'] == 'deleted') throw _conflict('deleted accounts cannot be changed');
    if (u['id'] == admin.id) throw _conflict('administrators cannot change their own role or status');
    final meta = <String, Object?>{};
    if (role != null && role != u['role']) {
      if (u['role'] == 'partner') throw _conflict('partner accounts are managed under /admin/v1/partners');
      meta['role'] = [u['role'], role];
      u['role'] = role;
    }
    if (status != null && status != u['status']) {
      meta['status'] = [u['status'], status];
      u['status'] = status;
    }
    if (meta.isNotEmpty) _audit(admin, 'admin.user.update', 'user', id, meta);
    return u;
  }

  Json _aPartners(Map<String, String?> q) {
    final st = q['status'];
    if (st != null && st.isNotEmpty && !const ['pending', 'approved', 'rejected', 'suspended'].contains(st)) throw _inv('status', 'invalid status');
    final list = _partners.where((p) => (st == null || st.isEmpty || p['status'] == st) && _has(p, 'businessName', q['q'])).toList().reversed.toList();
    final page = _pageOf(list, q);
    return {...page, 'items': (page['items'] as List).cast<Json>().map((p) => _partnerView(p, true)).toList()};
  }

  static const _transitions = {
    'pending': ['approved', 'rejected'],
    'approved': ['suspended'],
    'suspended': ['approved'],
    'rejected': ['approved'],
  };

  Json _aPatchPartner(UserProfile admin, int id, Json b) {
    final p = _find(_partners, id);
    final status = b['status'], note = b['decisionNote'];
    final num? rate = b['commissionRate'] as num?;
    if (status == null && rate == null && note == null) throw _inv('body', 'status, commissionRate or decisionNote required');
    if (status != null && !const ['approved', 'rejected', 'suspended'].contains(status)) throw _inv('status', 'must be approved, rejected or suspended');
    if (rate != null && (rate < 0 || rate > 50)) throw _inv('commissionRate', 'must be between 0 and 50');
    if (note != null && '$note'.length > 500) throw _inv('decisionNote', 'max 500 characters');
    final meta = <String, Object?>{};
    if (note != null) p['decisionNote'] = '$note'.trim();
    if (rate != null && _r2(rate.toDouble()) != p['commissionRate']) {
      meta['commissionRate'] = [p['commissionRate'], _r2(rate.toDouble())];
      p['commissionRate'] = _r2(rate.toDouble());
    }
    if (status != null && status != p['status']) {
      if (!(_transitions[p['status']] ?? const []).contains(status)) throw _conflict('cannot change partner from ${p['status']} to $status');
      if (status != 'approved' && '${p['decisionNote']}'.isEmpty) throw _inv('decisionNote', 'a note is required to reject or suspend');
      if (status == 'approved') {
        final owner = _find(_users, p['userId'] as int);
        if (owner['role'] != 'traveler' && owner['role'] != 'partner') throw _conflict('the owner account has role ${owner['role']} and cannot become a partner');
        owner['role'] = 'partner';
      }
      meta['status'] = [p['status'], status];
      p['status'] = status;
      p['decidedAt'] = _now().toIso8601String();
    }
    _audit(admin, 'admin.partner.update', 'partner', id, meta);
    return _partnerView(p, true);
  }

  Json _aListings(Map<String, String?> q) {
    final st = q['status'];
    if (st != null && st.isNotEmpty && !const ['draft', 'published', 'archived'].contains(st)) throw _inv('status', 'invalid status');
    final pid = int.tryParse(q['partnerId'] ?? '');
    final items = _listings
        .where((l) => (st == null || st.isEmpty || l['status'] == st) && (q['type'] == null || q['type']!.isEmpty || l['type'] == q['type']) && (pid == null || l['partnerId'] == pid) && (_has(l, 'name', q['q']) || _has(l, 'location', q['q'])))
        .map(_listingOut)
        .toList();
    return {'items': items};
  }

  Json _aPatchListing(UserProfile admin, int id, Json b) {
    final st = b['status'];
    if (st is! String || !const ['draft', 'published', 'archived'].contains(st)) throw _inv('status', 'must be draft, published or archived');
    final l = _find(_listings, id);
    final old = l['status'];
    l['status'] = st;
    _audit(admin, 'admin.listing.status', 'listing', id, {'status': [old, st]});
    return _listingOut(l);
  }

  Json _aAudit(Map<String, String?> q) {
    final a = q['action'];
    final actor = int.tryParse(q['actorId'] ?? '');
    final list = _audits.where((x) => (a == null || a.isEmpty || '${x['action']}'.startsWith(a)) && (actor == null || x['actorId'] == actor)).toList();
    return _pageOf(list, q);
  }

  // ---------- staff ----------
  Json _staffBooking(Json b) {
    final checks = List<Object?>.of(b['checks'] as List).cast<Json>().reversed.toList();
    return {
      for (final e in b.entries)
        if (e.key != 'lines' && e.key != 'userId') e.key: e.value,
      'checks': checks,
      'verification': checks.isEmpty ? 'none' : checks.first['result'],
    };
  }

  Json _sBookings(Map<String, String?> q) {
    final st = q['status'], checked = q['checked'];
    if (st != null && st.isNotEmpty && !const ['pending_payment', 'confirmed', 'completed', 'cancelled'].contains(st)) throw _inv('status', 'invalid status');
    if (checked != null && checked.isNotEmpty && checked != 'true' && checked != 'false') throw _inv('checked', 'must be true or false');
    final list = _bookings
        .where((b) =>
            (st == null || st.isEmpty || b['status'] == st) &&
            (checked == null || checked.isEmpty || (b['checks'] as List).isNotEmpty == (checked == 'true')) &&
            (_has(b, 'bookingNumber', q['q']) || _has(b, 'title', q['q'])))
        .toList()
      ..sort((a, c) => '${c['date']}'.compareTo('${a['date']}'));
    final page = _pageOf(list, q);
    return {...page, 'items': (page['items'] as List).cast<Json>().map(_staffBooking).toList()};
  }

  Json _sCheck(UserProfile u, int id, Json b) {
    final result = b['result'], note = '${b['note'] ?? ''}'.trim();
    if (result != 'verified' && result != 'flagged') throw _inv('result', 'must be verified or flagged');
    if (result == 'flagged' && note.isEmpty) throw _inv('note', 'required when flagging a booking');
    if (note.length > 500) throw _inv('note', 'max 500 characters');
    final bk = _find(_bookings, id);
    if (bk['status'] != 'confirmed' && bk['status'] != 'completed') throw _conflict('only paid bookings (confirmed or completed) are verified; this one is ${bk['status']}');
    (bk['checks'] as List).add({'id': _nextCheck++, 'staffName': _users.firstWhere((e) => e['id'] == u.id)['name'], 'result': result, 'note': note, 'createdAt': _now().toIso8601String()});
    _audit(u, 'staff.booking.check', 'booking', id, {'result': result});
    return _staffBooking(bk);
  }

  Json _staffTicket(Json t) {
    final u = _users.where((e) => e['id'] == t['userId']).firstOrNull;
    final replies = List<Object?>.of(t['replies'] as List);
    return {..._ticketOut(t), 'travelerName': u?['name'] ?? '', 'travelerEmail': u?['email'] ?? '', 'replyCount': replies.length};
  }

  Json _sTickets(Map<String, String?> q) {
    final st = q['status'];
    if (st != null && st.isNotEmpty && !const ['open', 'answered', 'resolved'].contains(st)) throw _inv('status', 'invalid status');
    final list = _tickets.where((t) => (st == null || st.isEmpty || t['status'] == st) && (_has(t, 'subject', q['q']) || _has(t, 'message', q['q']))).toList().reversed.toList();
    final page = _pageOf(list, q);
    return {...page, 'items': (page['items'] as List).cast<Json>().map(_staffTicket).toList()};
  }

  String _msg(Json b) {
    final m = '${b['message'] ?? ''}'.trim();
    if (m.isEmpty || m.length > 2000) throw _inv('message', 'required, max 2000 characters');
    return m;
  }

  Json _sReply(UserProfile u, int id, Json b) {
    final m = _msg(b);
    final t = _find(_tickets, id);
    if (t['status'] == 'resolved') throw _conflict('ticket is resolved; reopen it before replying');
    final who = _users.firstWhere((e) => e['id'] == u.id);
    (t['replies'] as List).add(_reply(_nextReply++, who['role'] == 'admin' ? 'admin' : 'staff', '${who['name']}', m, _now()));
    t['status'] = 'answered';
    t['firstRepliedAt'] ??= _now().toIso8601String();
    _audit(u, 'staff.ticket.reply', 'ticket', id);
    return _staffTicket(t);
  }

  Json _travelerReply(UserProfile u, int id, Json b) {
    final m = _msg(b);
    final t = _find(_tickets, id);
    if (t['userId'] != u.id) throw _notFound();
    (t['replies'] as List).add(_reply(_nextReply++, 'traveler', u.fullName, m, _now()));
    t['status'] = 'open';
    return _ticketOut(t);
  }

  Json _sPatchTicket(UserProfile u, int id, Json b) {
    final st = b['status'];
    if (st != 'open' && st != 'resolved') throw _inv('status', 'must be open or resolved');
    final t = _find(_tickets, id);
    if (t['status'] != st) {
      _audit(u, 'staff.ticket.status', 'ticket', id, {'status': [t['status'], st]});
      t['status'] = st;
      t['resolvedAt'] = st == 'resolved' ? _now().toIso8601String() : null;
    }
    return _staffTicket(t);
  }

  bool _onDay(Object? iso, DateTime from, DateTime to) {
    if (iso is! String) return false;
    final d = DateTime.tryParse(iso);
    return d != null && !d.isBefore(from) && d.isBefore(to);
  }

  Json _sCounters(DateTime from) {
    final to = from.add(const Duration(days: 1));
    final checks = [for (final b in _bookings) ...(b['checks'] as List).cast<Json>().where((c) => _onDay(c['createdAt'], from, to))];
    var sum = 0.0;
    var n = 0;
    for (final t in _tickets) {
      final f = t['firstRepliedAt'];
      if (f is String) {
        sum += DateTime.parse(f).difference(DateTime.parse('${t['createdAt']}')).inMinutes / 60;
        n++;
      }
    }
    final verified = checks.where((c) => c['result'] == 'verified').length;
    return {
      'openTickets': _tickets.where((t) => t['status'] == 'open').length,
      'uncheckedBookings': _bookings.where((b) => (b['status'] == 'confirmed' || b['status'] == 'completed') && (b['checks'] as List).isEmpty).length,
      'checkedToday': checks.length,
      'verifiedToday': verified,
      'flaggedToday': checks.length - verified,
      'avgFirstResponseHours': n == 0 ? 0.0 : _r2(sum / n),
    };
  }

  Json _sReport(UserProfile u, Map<String, String?> q) {
    var day = _startOfDay(_now());
    final d = q['date'];
    if (d != null && d.isNotEmpty) {
      final p = DateTime.tryParse(d);
      if (p == null || d.length != 10) throw _inv('date', 'must be YYYY-MM-DD');
      day = DateTime(p.year, p.month, p.day);
    }
    final end = day.add(const Duration(days: 1));
    final c = _sCounters(day);
    final who = _users.firstWhere((e) => e['id'] == u.id)['name'];
    final mine = [for (final b in _bookings) ...(b['checks'] as List).cast<Json>().where((x) => x['staffName'] == who && _onDay(x['createdAt'], day, end))];
    return {
      'date': '${day.year}-${day.month.toString().padLeft(2, '0')}-${day.day.toString().padLeft(2, '0')}',
      'timezone': 'Asia/Bangkok',
      'bookingsCreated': _bookings.where((b) => _onDay(b['date'], day, end)).length,
      'bookingsVerified': c['verifiedToday'],
      'bookingsFlagged': c['flaggedToday'],
      'myChecks': mine.length,
      'ticketsOpened': _tickets.where((t) => _onDay(t['createdAt'], day, end)).length,
      'ticketsFirstAnswered': _tickets.where((t) => _onDay(t['firstRepliedAt'], day, end)).length,
      'ticketsResolved': _tickets.where((t) => _onDay(t['resolvedAt'], day, end)).length,
    };
  }
}
