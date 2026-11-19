import 'package:flutter/material.dart';

import '../core/responsive.dart';
import '../core/theme.dart';
import 'bookings_screen.dart';
import 'home_screen.dart';
import 'profile_screen.dart';
import 'search_screen.dart';
import 'wallet_screen.dart';
import 'widgets/brand.dart';

/// Search request pushed from Home (category chip / destination card) into the Search tab.
class SearchIntent {
  const SearchIntent(
      {this.type, this.destinationId, this.destinationName, this.query});
  final String? type, destinationName, query;
  final int? destinationId;
}

/// Bottom navigation: Home · Search · Bookings · Wallet · Profile (board main menu, Sprint 1 subset).
class AppShell extends StatefulWidget {
  const AppShell({super.key, this.modeLabel = 'mock'});
  final String modeLabel;
  @override
  State<AppShell> createState() => _AppShellState();
}

class _AppShellState extends State<AppShell> {
  int _tab = 0;
  int _searchVersion = 0;
  SearchIntent _intent = const SearchIntent();

  int _bookingsVersion = 0, _walletVersion = 0;

  /// Bottom-nav / main-menu navigation. Bookings and Wallet are rebuilt (and therefore reloaded) whenever the
  /// user switches to them, because IndexedStack keeps tabs alive and their data changes after payments.
  void _goTab(int i) => setState(() {
        if (i == 1 && _tab != 1) _intent = const SearchIntent();
        if (i == 2 && _tab != 2) _bookingsVersion++;
        if (i == 3 && _tab != 3) _walletVersion++;
        _tab = i;
      });

  void _openSearch(SearchIntent i) => setState(() {
        _intent = i;
        _searchVersion++;
        _tab = 1;
      });

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    final extendedRail = width >= TepBreakpoints.expanded;

    final pages = <Widget>[
      HomeScreen(
          onSearch: _openSearch, onTab: _goTab, modeLabel: widget.modeLabel),
      SearchScreen(key: ValueKey('search-$_searchVersion'), intent: _intent),
      BookingsScreen(key: ValueKey('bookings-$_bookingsVersion')),
      WalletScreen(key: ValueKey('wallet-$_walletVersion')),
      const ProfileScreen(),
    ];
    final useRail = TepBreakpoints.isExpanded(context);
    return Scaffold(
      body: Row(children: [
        if (useRail)
          NavigationRail(
            key: const Key('desktop-navigation'),
            selectedIndex: _tab,
            onDestinationSelected: _goTab,
            extended: extendedRail,
            minExtendedWidth: 220,
            leading: Padding(
              padding: const EdgeInsets.fromLTRB(12, 16, 12, 10),
              child: extendedRail
                  ? const TepBrandMark(compact: true)
                  : Container(
                      width: 44,
                      height: 44,
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                          color: TepColors.primaryLight,
                          borderRadius: BorderRadius.circular(14)),
                      child: const Text('TEP',
                          style: TextStyle(
                              color: TepColors.primaryDeep,
                              fontWeight: FontWeight.w900))),
            ),
            labelType: extendedRail
                ? NavigationRailLabelType.none
                : NavigationRailLabelType.all,
            backgroundColor: Colors.white,
            selectedIconTheme: const IconThemeData(color: TepColors.primary),
            selectedLabelTextStyle: const TextStyle(
                color: TepColors.primary, fontWeight: FontWeight.w800),
            unselectedLabelTextStyle: const TextStyle(
                color: TepColors.muted, fontWeight: FontWeight.w600),
            destinations: const [
              NavigationRailDestination(
                  icon: Icon(Icons.home_outlined),
                  selectedIcon: Icon(Icons.home_rounded),
                  label: Text('หน้าแรก')),
              NavigationRailDestination(
                  icon: Icon(Icons.search_rounded),
                  selectedIcon: Icon(Icons.search_rounded),
                  label: Text('ค้นหา')),
              NavigationRailDestination(
                  icon: Icon(Icons.receipt_long_outlined),
                  selectedIcon: Icon(Icons.receipt_long_rounded),
                  label: Text('การจอง')),
              NavigationRailDestination(
                  icon: Icon(Icons.account_balance_wallet_outlined),
                  selectedIcon: Icon(Icons.account_balance_wallet_rounded),
                  label: Text('Wallet')),
              NavigationRailDestination(
                  icon: Icon(Icons.person_outline_rounded),
                  selectedIcon: Icon(Icons.person_rounded),
                  label: Text('โปรไฟล์')),
            ],
          ),
        Expanded(
            child: TepConstrainedContent(
                maxWidth: useRail ? 1200 : double.infinity,
                child: IndexedStack(index: _tab, children: pages))),
      ]),
      bottomNavigationBar: useRail
          ? null
          : NavigationBar(
              selectedIndex: _tab,
              onDestinationSelected: _goTab,
              destinations: const [
                NavigationDestination(
                    icon: Icon(Icons.home_outlined),
                    selectedIcon: Icon(Icons.home_rounded),
                    label: 'หน้าแรก'),
                NavigationDestination(
                    icon: Icon(Icons.search_rounded),
                    selectedIcon: Icon(Icons.search_rounded),
                    label: 'ค้นหา'),
                NavigationDestination(
                    icon: Icon(Icons.receipt_long_outlined),
                    selectedIcon: Icon(Icons.receipt_long_rounded),
                    label: 'การจอง'),
                NavigationDestination(
                    icon: Icon(Icons.account_balance_wallet_outlined),
                    selectedIcon: Icon(Icons.account_balance_wallet_rounded),
                    label: 'Wallet'),
                NavigationDestination(
                    icon: Icon(Icons.person_outline_rounded),
                    selectedIcon: Icon(Icons.person_rounded),
                    label: 'โปรไฟล์'),
              ],
            ),
    );
  }
}
