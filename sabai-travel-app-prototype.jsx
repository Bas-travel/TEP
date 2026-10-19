import React, { useState, useRef, useEffect } from 'react';
import {
  Home, Search, Sparkles, Briefcase, User, Bell, MapPin, Star, ChevronRight,
  ArrowLeft, Calendar, Wallet, CreditCard, CheckCircle2, Plane, Hotel,
  UtensilsCrossed, Coffee, SlidersHorizontal, Plus, Minus, Clock, Mail, Lock,
  ChevronDown, Heart, LogOut, Settings, HelpCircle, Ticket, Wifi, Car, Waves,
  Landmark, QrCode, FileText, Phone, ArrowDown, ArrowUp, Copy, Check,
  MessageCircle, ChevronUp, Send, AlertCircle, Globe, Banknote, Palette, KeyRound, Trash2,
  Users, UserPlus, UserCheck, Bookmark, Share2, MoreHorizontal, Flag, EyeOff, BookOpen, Lightbulb,
  Navigation, Map as MapIcon, Route as RouteIcon, X
} from 'lucide-react';

/* ---------------------------------------------------------------- */
/*  THEME + FONTS                                                    */
/* ---------------------------------------------------------------- */

const FONT_CSS = `
@import url('https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,400;9..144,500;9..144,600;9..144,700&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap');
:root{
  --ink:#16241F; --ink-soft:#5C6A64; --paper:#FAFAF8; --surface:#FFFFFF;
  --pine:#10403D; --pine-light:#1C5C57; --pine-pale:#E4EDEA;
  --amber:#E3A23C; --amber-soft:#F6E4C1; --border:#E4E1D8; --danger:#C0483A;
  --amber-ink:#8A5A17; --on-amber:#4A2E06; --on-amber-solid:#4A2E06; --danger-soft:#F6DCD6; --muted-bg:#EDEAE2; --disabled:#B9C4C1;
}
.theme-dark{
  --ink:#EDF1EE; --ink-soft:#9FB2AB; --paper:#0E1614; --surface:#17231F;
  --pine:#2F8A7F; --pine-light:#3A9D91; --pine-pale:#1D312C;
  --amber:#E3A23C; --amber-soft:#3A2E17; --border:#2B3B36; --danger:#E0705F;
  --amber-ink:#F0B95C; --on-amber:#F6E4C1; --on-amber-solid:#4A2E06; --danger-soft:#3C201C; --muted-bg:#23312D; --disabled:#3B4B46;
  color-scheme:dark;
}
.font-display{ font-family:'Fraunces', serif; }
@keyframes spin-slow{ to{ transform:rotate(360deg); } }
@keyframes fade-up{ from{opacity:0; transform:translateY(8px);} to{opacity:1; transform:translateY(0);} }
.fade-up{ animation: fade-up .35s ease both; }
.scrollbar-hide::-webkit-scrollbar{ display:none; }
.scrollbar-hide{ -ms-overflow-style:none; scrollbar-width:none; }
`;

/* ---------------------------------------------------------------- */
/*  MOCK DATA                                                         */
/* ---------------------------------------------------------------- */

const DESTINATIONS = [
  { id: 'chiangmai', name: 'เชียงใหม่', sub: 'ภูเขา ธรรมชาติ วัฒนธรรมล้านนา', gradient: 'linear-gradient(135deg,#1C5C57,#89A66B)' },
  { id: 'phuket', name: 'ภูเก็ต', sub: 'ทะเลอันดามัน ชายหาดสวย', gradient: 'linear-gradient(135deg,#0B4F6C,#01BAEF)' },
  { id: 'krabi', name: 'กระบี่', sub: 'หน้าผาหินปูน เกาะแก่ง', gradient: 'linear-gradient(135deg,#123C69,#3E92CC)' },
  { id: 'pai', name: 'ปาย', sub: 'หมอกยามเช้า คาเฟ่ริมเขา', gradient: 'linear-gradient(135deg,#7A4B22,#E3A23C)' },
  { id: 'bangkok', name: 'กรุงเทพฯ', sub: 'เมืองหลวง สตรีทฟู้ด ช้อปปิ้ง', gradient: 'linear-gradient(135deg,#3B2352,#A768C9)' },
  { id: 'samui', name: 'เกาะสมุย', sub: 'เกาะสวรรค์ มะพร้าว ทะเลใส', gradient: 'linear-gradient(135deg,#0E6E55,#59C9A5)' },
];

const LISTINGS = [
  { id: 1, type: 'hotel', name: 'ล้านนา วิว รีสอร์ท', location: 'นิมมานเหมินท์, เชียงใหม่', rating: 4.8, reviews: 342, price: 1490 },
  { id: 2, type: 'restaurant', name: 'ข้าวซอยแม่สาย', location: 'เมืองเก่า, เชียงใหม่', rating: 4.6, reviews: 891, price: 120 },
  { id: 3, type: 'attraction', name: 'วัดพระธาตุดอยสุเทพ', location: 'ดอยสุเทพ, เชียงใหม่', rating: 4.9, reviews: 5210, price: 50 },
  { id: 4, type: 'hotel', name: 'อันดามัน บีช วิลลา', location: 'ป่าตอง, ภูเก็ต', rating: 4.7, reviews: 206, price: 2350 },
  { id: 5, type: 'flight', name: 'กรุงเทพฯ → เชียงใหม่', location: 'สายการบิน Sabai Air', rating: 4.4, reviews: 1032, price: 990 },
  { id: 6, type: 'restaurant', name: 'ครัวทะเลกระบี่', location: 'อ่าวนาง, กระบี่', rating: 4.5, reviews: 410, price: 250 },
];

const STYLE_OPTIONS = ['ธรรมชาติ', 'อาหารท้องถิ่น', 'วัฒนธรรม', 'ผจญภัย', 'ช้อปปิ้ง', 'พักผ่อนสบายๆ'];
const TRAVEL_WITH_OPTIONS = ['คนเดียว', 'คู่รัก', 'ครอบครัว', 'เพื่อน', 'กลุ่มทัวร์'];
const BUDGET_LEVELS = ['ประหยัด', 'ปานกลาง', 'หรูหรา'];

const TYPE_ICON = { hotel: Hotel, restaurant: UtensilsCrossed, flight: Plane, attraction: Landmark, trip: Sparkles };
const TYPE_GRADIENT = {
  hotel: 'linear-gradient(135deg,#1C5C57,#0B2422)',
  restaurant: 'linear-gradient(135deg,#B5651D,#7A4B22)',
  flight: 'linear-gradient(135deg,#123C69,#3E92CC)',
  attraction: 'linear-gradient(135deg,#3B2352,#A768C9)',
  trip: 'linear-gradient(135deg,#10403D,#E3A23C)',
};
const TYPE_LABEL = { hotel: 'ที่พัก', restaurant: 'อาหาร', flight: 'เที่ยวบิน', attraction: 'สถานที่เที่ยว', trip: 'ทริป' };

const NOTIF_ICON = { booking: FileText, payment: CreditCard, promotion: Ticket, trip: Briefcase, system: Settings, community: Users };

const INITIAL_COUPONS = [
  { id: 1, code: 'SABAI20', title: 'ลด 20% ทริปแรก', detail: 'ใช้ได้กับการจองที่พักและเที่ยวบินทุกรายการ', minSpend: 1000, maxDiscount: 500, expiry: '31 ต.ค. 2569', status: 'available' },
  { id: 2, code: 'HOTEL150', title: 'ลด ฿150 ที่พัก', detail: 'สำหรับการจองที่พักขั้นต่ำ ฿1,500 ขึ้นไป', minSpend: 1500, maxDiscount: 150, expiry: '15 พ.ย. 2569', status: 'available' },
  { id: 3, code: 'FOOD50', title: 'ลด ฿50 ร้านอาหาร', detail: 'ใช้ได้กับร้านอาหารที่ร่วมรายการ', minSpend: 200, maxDiscount: 50, expiry: '20 ต.ค. 2569', status: 'available' },
  { id: 4, code: 'WELCOME100', title: 'ลด ฿100 สมาชิกใหม่', detail: 'สำหรับสมาชิกใหม่ในการจองครั้งแรก', minSpend: 500, maxDiscount: 100, expiry: '5 ก.ย. 2569', status: 'used' },
  { id: 5, code: 'FLASH30', title: 'ลด 30% แฟลชเซล', detail: 'โปรโมชั่นพิเศษช่วงเทศกาล', minSpend: 1000, maxDiscount: 800, expiry: '1 ส.ค. 2569', status: 'expired' },
];

const FAQS = [
  { id: 1, q: 'ยกเลิกการจองได้อย่างไร?', a: 'ไปที่ "ทริป" หรือ "การจองของฉัน" เลือกรายการที่ต้องการยกเลิก แล้วกด "ยกเลิกการจอง" ระบบจะคืนเงินตามเงื่อนไขของแต่ละที่พัก/ร้านค้า' },
  { id: 2, q: 'เงินคืนใช้เวลากี่วัน?', a: 'เงินคืนเข้า Sabai Wallet ทันทีหลังอนุมัติ หรือเข้าบัญชีธนาคาร/บัตรเครดิตภายใน 5-10 วันทำการ ขึ้นอยู่กับวิธีชำระเงินเดิม' },
  { id: 3, q: 'ใช้คูปองส่วนลดอย่างไร?', a: 'กรอกรหัสคูปองในขั้นตอน "สรุปการจอง" ก่อนชำระเงิน หรือเลือกจากหน้า "คูปองของฉัน" แล้วกดใช้งาน ระบบจะคำนวณส่วนลดให้อัตโนมัติ' },
  { id: 4, q: 'เปลี่ยนวันเดินทางได้ไหม?', a: 'ขึ้นอยู่กับนโยบายของแต่ละที่พัก/สายการบิน บางรายการเปลี่ยนแปลงได้ฟรี บางรายการมีค่าธรรมเนียม ตรวจสอบได้ที่หน้ารายละเอียดการจอง' },
  { id: 5, q: 'ติดต่อฝ่ายสนับสนุนได้ช่องทางไหนบ้าง?', a: 'แชทผ่านแอปได้ตลอด 24 ชั่วโมง หรือสร้างตั๋วสนับสนุน (Support Ticket) ทีมงานจะตอบกลับภายใน 24 ชั่วโมง' },
];

function buildItinerary({ destinationName, days, budget, styles }) {
  if (destinationName === 'เชียงใหม่' && days === 3) {
    return {
      days: [
        { title: 'วันที่ 1 · เมืองเก่าเชียงใหม่', items: [
          { time: '09:00', text: 'เดินทางถึงสนามบินเชียงใหม่ เช็คอินที่พัก', place: 'สนามบินเชียงใหม่' },
          { time: '12:00', text: 'อาหารกลางวันข้าวซอยที่ร้านเมืองเก่า', place: 'ข้าวซอย เมืองเก่า เชียงใหม่' },
          { time: '14:00', text: 'เที่ยววัดพระสิงห์และเดินเล่นถนนคนเดินท่าแพ', place: 'วัดพระสิงห์ เชียงใหม่' },
          { time: '18:30', text: 'อาหารเย็นที่ตลาดวโรรส', place: 'ตลาดวโรรส เชียงใหม่' },
        ]},
        { title: 'วันที่ 2 · ดอยอินทนนท์', items: [
          { time: '07:00', text: 'ออกเดินทางสู่ดอยอินทนนท์' },
          { time: '09:30', text: 'ชมพระธาตุนภเมทนีดลและนภพลภูมิสิริ', place: 'พระมหาธาตุนภเมทนีดล ดอยอินทนนท์' },
          { time: '12:00', text: 'ปิคนิคอาหารกลางวันริมธารน้ำ' },
          { time: '15:00', text: 'เดินชมน้ำตกวชิรธาร', place: 'น้ำตกวชิรธาร ดอยอินทนนท์' },
          { time: '19:00', text: 'กลับเข้าเมือง อาหารเย็นชิลๆ' },
        ]},
        { title: 'วันที่ 3 · คาเฟ่และตลาด', items: [
          { time: '09:00', text: 'จิบกาแฟที่คาเฟ่ย่านนิมมานเหมินท์', place: 'คาเฟ่ นิมมานเหมินท์ เชียงใหม่' },
          { time: '11:00', text: 'ช้อปปิ้งที่ One Nimman', place: 'One Nimman เชียงใหม่' },
          { time: '14:00', text: 'เดินเล่นตลาดเมืองใหม่', place: 'ตลาดเมืองใหม่ เชียงใหม่' },
          { time: '17:00', text: 'เดินทางกลับสนามบิน', place: 'สนามบินเชียงใหม่' },
        ]},
      ],
      budgetBreakdown: [
        { label: 'เที่ยวบิน', value: 3000 },
        { label: 'ที่พัก', value: 4000 },
        { label: 'อาหาร', value: 1500 },
        { label: 'กิจกรรม', value: 1000 },
        { label: 'การเดินทางในพื้นที่', value: 500 },
      ],
    };
  }
  const templates = [
    (d) => ({ text: `สำรวจแหล่งท่องเที่ยวสำคัญของ${d}`, place: `สถานที่ท่องเที่ยว ${d}` }),
    (d) => ({ text: `ลิ้มลองอาหารท้องถิ่นชื่อดังของ${d}`, place: `อาหารท้องถิ่น ${d}` }),
    (d) => ({ text: 'เดินเล่นตลาดและย่านคาเฟ่ยอดนิยม', place: `ตลาด คาเฟ่ ${d}` }),
    (d) => ({ text: `กิจกรรมกลางแจ้งท่ามกลางธรรมชาติใกล้${d}`, place: `ธรรมชาติ จุดชมวิว ${d}` }),
    (d) => ({ text: 'พักผ่อนและช้อปปิ้งของฝาก', place: `ของฝาก ${d}` }),
  ];
  const dayList = Array.from({ length: Math.min(days, 7) }).map((_, i) => {
    const first = templates[i % templates.length](destinationName);
    return {
      title: `วันที่ ${i + 1} · ${destinationName}`,
      items: [
        { time: '09:00', ...first },
        { time: '13:00', text: `อาหารกลางวันร้านแนะนำใน${destinationName}`, place: `ร้านอาหาร ${destinationName}` },
        { time: '16:00', text: styles.length ? `กิจกรรมตามความสนใจ: ${styles.join(', ')}` : 'เดินเล่นชมเมือง', place: `${styles[0] || 'สถานที่ท่องเที่ยว'} ${destinationName}` },
        { time: '19:00', text: 'อาหารเย็นและพักผ่อน', place: `ร้านอาหารมื้อเย็น ${destinationName}` },
      ],
    };
  });
  const b = Math.max(budget, 3000);
  return {
    days: dayList,
    budgetBreakdown: [
      { label: 'เที่ยวบิน', value: Math.round((b * 0.28) / 100) * 100 },
      { label: 'ที่พัก', value: Math.round((b * 0.35) / 100) * 100 },
      { label: 'อาหาร', value: Math.round((b * 0.2) / 100) * 100 },
      { label: 'กิจกรรม', value: Math.round((b * 0.1) / 100) * 100 },
      { label: 'การเดินทางในพื้นที่', value: Math.round((b * 0.07) / 100) * 100 },
    ],
  };
}

/* ---------------------------------------------------------------- */
/*  SHARED UI PIECES                                                  */
/* ---------------------------------------------------------------- */

function TopBar({ title, onBack, right }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '14px 20px', borderBottom: '1px solid var(--border)', background: 'var(--surface)', flexShrink: 0 }}>
      {onBack && (
        <button onClick={onBack} style={{ background: 'var(--pine-pale)', border: 'none', width: 36, height: 36, borderRadius: 999, display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}>
          <ArrowLeft size={18} color="var(--pine)" />
        </button>
      )}
      <h1 className="font-display" style={{ fontSize: 18, fontWeight: 600, color: 'var(--ink)', flex: 1 }}>{title}</h1>
      {right}
    </div>
  );
}

function TabBar({ active, onChange }) {
  const tabs = [
    { id: 'home', label: 'หน้าแรก', icon: Home },
    { id: 'search', label: 'ค้นหา', icon: Search },
    { id: 'ai', label: 'AI', icon: Sparkles },
    { id: 'trips', label: 'ทริป', icon: Briefcase },
    { id: 'community', label: 'ชุมชน', icon: Users },
    { id: 'profile', label: 'โปรไฟล์', icon: User },
  ];
  return (
    <div style={{ display: 'flex', borderTop: '1px solid var(--border)', background: 'var(--surface)', flexShrink: 0 }}>
      {tabs.map((t) => {
        const Icon = t.icon;
        const isActive = active === t.id;
        return (
          <button key={t.id} onClick={() => onChange(t.id)} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4, padding: '10px 0 14px', background: 'transparent', border: 'none', cursor: 'pointer' }}>
            <Icon size={22} strokeWidth={isActive ? 2.4 : 1.8} color={isActive ? 'var(--pine)' : 'var(--ink-soft)'} />
            <span style={{ fontSize: 11, fontWeight: isActive ? 700 : 500, color: isActive ? 'var(--pine)' : 'var(--ink-soft)' }}>{t.label}</span>
          </button>
        );
      })}
    </div>
  );
}

function Chip({ label, active, onClick }) {
  return (
    <button onClick={onClick} style={{
      padding: '8px 16px', borderRadius: 999, fontSize: 13, fontWeight: 600, whiteSpace: 'nowrap',
      border: active ? '1.5px solid var(--pine)' : '1.5px solid var(--border)',
      background: active ? 'var(--pine)' : 'var(--surface)',
      color: active ? '#fff' : 'var(--ink)', cursor: 'pointer',
    }}>{label}</button>
  );
}

function Stepper({ value, onChange, min = 1, max = 14 }) {
  const btn = { width: 36, height: 36, borderRadius: 999, border: '1.5px solid var(--border)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' };
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
      <button onClick={() => onChange(Math.max(min, value - 1))} style={btn}><Minus size={16} color="var(--ink)" /></button>
      <span className="font-display" style={{ fontSize: 20, fontWeight: 600, minWidth: 28, textAlign: 'center' }}>{value}</span>
      <button onClick={() => onChange(Math.min(max, value + 1))} style={btn}><Plus size={16} color="var(--ink)" /></button>
    </div>
  );
}

function RatingStars({ value }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 3 }}>
      <Star size={13} fill="var(--amber)" color="var(--amber)" />
      <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--ink)' }}>{value}</span>
    </div>
  );
}

function TextField({ icon: Icon, type = 'text', placeholder, value, onChange, disabled }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10, border: '1.5px solid var(--border)', borderRadius: 14, padding: '12px 14px', background: 'var(--surface)', opacity: disabled ? 0.55 : 1 }}>
      {Icon && <Icon size={18} color="var(--ink-soft)" />}
      <input type={type} placeholder={placeholder} value={value} onChange={onChange} disabled={disabled}
        style={{ border: 'none', outline: 'none', flex: 1, fontSize: 14, fontFamily: "'Plus Jakarta Sans', sans-serif", background: 'transparent', color: 'var(--ink)', width: '100%' }} />
    </div>
  );
}

function PrimaryButton({ children, onClick, disabled }) {
  return (
    <button onClick={onClick} disabled={disabled} style={{ width: '100%', padding: '15px', borderRadius: 16, border: 'none', background: disabled ? 'var(--disabled)' : 'var(--pine)', color: '#fff', fontWeight: 700, fontSize: 15, cursor: disabled ? 'default' : 'pointer' }}>
      {children}
    </button>
  );
}

function SecondaryButton({ children, onClick }) {
  return (
    <button onClick={onClick} style={{ width: '100%', padding: '15px', borderRadius: 16, border: '1.5px solid var(--pine)', background: 'transparent', color: 'var(--pine)', fontWeight: 700, fontSize: 15, cursor: 'pointer' }}>
      {children}
    </button>
  );
}

function IconCircle({ icon: Icon, bg, color, onClick, label }) {
  return (
    <button onClick={onClick} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6, background: 'transparent', border: 'none', cursor: 'pointer' }}>
      <div style={{ width: 52, height: 52, borderRadius: 18, background: bg, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Icon size={22} color={color} />
      </div>
      <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--ink)' }}>{label}</span>
    </button>
  );
}

function StepDots({ current, total }) {
  return (
    <div style={{ display: 'flex', gap: 6 }}>
      {Array.from({ length: total }).map((_, i) => (
        <div key={i} style={{ flex: 1, height: 4, borderRadius: 999, background: i < current ? 'var(--pine)' : 'var(--border)' }} />
      ))}
    </div>
  );
}

/* ---------------------------------------------------------------- */
/*  SCREENS                                                           */
/* ---------------------------------------------------------------- */

function LoginScreen({ email, setEmail, password, setPassword, onLogin, onGuest, onRegister }) {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div style={{ background: 'linear-gradient(160deg, var(--pine) 0%, #0B2422 100%)', padding: '52px 28px 40px', color: '#fff', flexShrink: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 24 }}>
          <div style={{ width: 34, height: 34, borderRadius: 10, background: 'var(--amber)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <Sparkles size={18} color="#10403D" />
          </div>
          <span style={{ fontWeight: 700, letterSpacing: 0.2 }}>Sabai</span>
        </div>
        <h1 className="font-display" style={{ fontSize: 32, lineHeight: 1.2, fontWeight: 600, marginBottom: 10 }}>เที่ยวให้สบาย<br />ครบในแอปเดียว</h1>
        <p style={{ fontSize: 13.5, opacity: 0.8, maxWidth: 260 }}>ค้นหา วางแผน จอง และเดินทาง — ให้ AI ช่วยจัดทริปของคุณตั้งแต่ต้นจนจบ</p>
      </div>
      <div className="fade-up scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', background: 'var(--surface)', borderTopLeftRadius: 28, borderTopRightRadius: 28, marginTop: -22, padding: '26px 24px', display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--ink)', marginBottom: 6 }}>อีเมลหรือเบอร์โทรศัพท์</p>
          <TextField icon={Mail} placeholder="you@email.com" value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--ink)', marginBottom: 6 }}>รหัสผ่าน</p>
          <TextField icon={Lock} type="password" placeholder="••••••••" value={password} onChange={(e) => setPassword(e.target.value)} />
        </div>
        <button style={{ alignSelf: 'flex-end', background: 'none', border: 'none', color: 'var(--pine)', fontSize: 12, fontWeight: 600, cursor: 'pointer' }}>ลืมรหัสผ่าน?</button>
        <div style={{ marginTop: 4 }}><PrimaryButton onClick={onLogin}>เข้าสู่ระบบ</PrimaryButton></div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, margin: '4px 0' }}>
          <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
          <span style={{ fontSize: 12, color: 'var(--ink-soft)' }}>หรือ</span>
          <div style={{ flex: 1, height: 1, background: 'var(--border)' }} />
        </div>
        <button onClick={onGuest} style={{ background: 'none', border: 'none', color: 'var(--ink-soft)', fontSize: 13, fontWeight: 600, textDecoration: 'underline', cursor: 'pointer' }}>เข้าใช้งานแบบผู้เยี่ยมชม</button>
        <p style={{ textAlign: 'center', fontSize: 12, color: 'var(--ink-soft)', marginTop: 14 }}>
          ยังไม่มีบัญชี?{' '}
          <span onClick={onRegister} style={{ color: 'var(--pine)', fontWeight: 700, cursor: 'pointer' }}>สมัครสมาชิก</span>
        </p>
      </div>
    </div>
  );
}

function RegisterScreen({ step, form, onChange, onToggleStyle, onBack, onNext, onFinish }) {
  const step1Valid = form.firstName && form.lastName && form.email && form.password && form.password === form.confirmPassword;
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="สมัครสมาชิก" onBack={onBack} />
      <div style={{ padding: '14px 20px 0', flexShrink: 0 }}>
        <StepDots current={step} total={3} />
        <p style={{ fontSize: 11, color: 'var(--ink-soft)', marginTop: 6 }}>ขั้นตอนที่ {step} จาก 3</p>
      </div>
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: '16px 20px 20px', display: 'flex', flexDirection: 'column', gap: 14 }}>
        {step === 1 && (
          <>
            <p className="font-display" style={{ fontSize: 19, fontWeight: 600 }}>ข้อมูลส่วนตัว</p>
            <div style={{ display: 'flex', gap: 10 }}>
              <div style={{ flex: 1 }}><TextField icon={User} placeholder="ชื่อจริง" value={form.firstName} onChange={(e) => onChange({ firstName: e.target.value })} /></div>
              <div style={{ flex: 1 }}><TextField icon={User} placeholder="นามสกุล" value={form.lastName} onChange={(e) => onChange({ lastName: e.target.value })} /></div>
            </div>
            <TextField icon={Mail} placeholder="อีเมล" value={form.email} onChange={(e) => onChange({ email: e.target.value })} />
            <TextField icon={Phone} placeholder="เบอร์โทรศัพท์" value={form.phone} onChange={(e) => onChange({ phone: e.target.value })} />
            <TextField icon={Lock} type="password" placeholder="รหัสผ่าน" value={form.password} onChange={(e) => onChange({ password: e.target.value })} />
            <TextField icon={Lock} type="password" placeholder="ยืนยันรหัสผ่าน" value={form.confirmPassword} onChange={(e) => onChange({ confirmPassword: e.target.value })} />
            {form.confirmPassword && form.password !== form.confirmPassword && (
              <p style={{ fontSize: 11.5, color: 'var(--danger)' }}>รหัสผ่านไม่ตรงกัน</p>
            )}
          </>
        )}
        {step === 2 && (
          <>
            <p className="font-display" style={{ fontSize: 19, fontWeight: 600 }}>ข้อมูลโปรไฟล์</p>
            <div>
              <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>เพศ</p>
              <div style={{ display: 'flex', gap: 8 }}>
                {['หญิง', 'ชาย', 'ไม่ระบุ'].map((g) => <Chip key={g} label={g} active={form.gender === g} onClick={() => onChange({ gender: g })} />)}
              </div>
            </div>
            <div>
              <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>วันเกิด</p>
              <TextField icon={Calendar} type="date" value={form.dob} onChange={(e) => onChange({ dob: e.target.value })} />
            </div>
            <div>
              <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>สัญชาติ</p>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {['ไทย', 'ต่างชาติ'].map((n) => <Chip key={n} label={n} active={form.nationality === n} onClick={() => onChange({ nationality: n })} />)}
              </div>
            </div>
            <div>
              <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>ภาษาที่ใช้งาน</p>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {['ไทย', 'English', '中文', '日本語'].map((l) => <Chip key={l} label={l} active={form.language === l} onClick={() => onChange({ language: l })} />)}
              </div>
            </div>
          </>
        )}
        {step === 3 && (
          <>
            <p className="font-display" style={{ fontSize: 19, fontWeight: 600 }}>ความชอบด้านการเที่ยว</p>
            <div>
              <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>งบประมาณที่สบายใจ</p>
              <div style={{ display: 'flex', gap: 8 }}>
                {BUDGET_LEVELS.map((b) => <Chip key={b} label={b} active={form.budgetLevel === b} onClick={() => onChange({ budgetLevel: b })} />)}
              </div>
            </div>
            <div>
              <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>สไตล์การเที่ยวที่ชอบ</p>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {STYLE_OPTIONS.map((s) => <Chip key={s} label={s} active={form.travelStyles.includes(s)} onClick={() => onToggleStyle(s)} />)}
              </div>
            </div>
            <div>
              <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>จุดหมายที่อยากไป</p>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {DESTINATIONS.map((d) => <Chip key={d.id} label={d.name} active={form.favoriteDestination === d.name} onClick={() => onChange({ favoriteDestination: d.name })} />)}
              </div>
            </div>
            <div>
              <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>ปกติเที่ยวกับใคร</p>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                {TRAVEL_WITH_OPTIONS.map((t) => <Chip key={t} label={t} active={form.travelWith === t} onClick={() => onChange({ travelWith: t })} />)}
              </div>
            </div>
          </>
        )}
      </div>
      <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        {step < 3 ? (
          <PrimaryButton onClick={onNext} disabled={step === 1 && !step1Valid}>ถัดไป</PrimaryButton>
        ) : (
          <PrimaryButton onClick={onFinish}>เริ่มต้นใช้งาน 🎉</PrimaryButton>
        )}
      </div>
    </div>
  );
}

function HomeScreen({ onSearchTap, onQuickAction, onAiTap, onDestinationTap, onNotificationsTap, unreadCount }) {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: '20px 20px 28px', display: 'flex', flexDirection: 'column', gap: 22 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--ink-soft)', fontSize: 12, fontWeight: 600 }}>
            <MapPin size={14} /> กรุงเทพฯ, ประเทศไทย <ChevronDown size={14} />
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <button onClick={onNotificationsTap} style={{ position: 'relative', background: 'none', border: 'none', cursor: 'pointer', padding: 0, display: 'flex' }}>
              <Bell size={20} color="var(--ink)" />
              {unreadCount > 0 && <span style={{ position: 'absolute', top: -2, right: -2, width: 8, height: 8, borderRadius: 999, background: 'var(--danger)' }} />}
            </button>
            <div style={{ width: 34, height: 34, borderRadius: 999, background: 'var(--amber-soft)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, color: 'var(--pine)', fontSize: 13 }}>น</div>
          </div>
        </div>

        <div>
          <p className="font-display" style={{ fontSize: 24, fontWeight: 600, color: 'var(--ink)' }}>สวัสดี, หนาน 👋</p>
          <p style={{ fontSize: 13, color: 'var(--ink-soft)', marginTop: 2 }}>วันนี้อยากไปเที่ยวที่ไหนดี?</p>
        </div>

        <button onClick={onSearchTap} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '14px 16px', borderRadius: 16, border: '1.5px solid var(--border)', background: 'var(--surface)', cursor: 'pointer' }}>
          <Search size={18} color="var(--ink-soft)" />
          <span style={{ fontSize: 13, color: 'var(--ink-soft)' }}>ค้นหาที่พัก ร้านอาหาร สถานที่เที่ยว</span>
        </button>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 6 }}>
          <IconCircle icon={Plane} bg="#E4EDEA" color="var(--pine)" label="เที่ยวบิน" onClick={() => onQuickAction('เที่ยวบิน')} />
          <IconCircle icon={Hotel} bg="#E4EDEA" color="var(--pine)" label="ที่พัก" onClick={() => onQuickAction('ที่พัก')} />
          <IconCircle icon={UtensilsCrossed} bg="#E4EDEA" color="var(--pine)" label="อาหาร" onClick={() => onQuickAction('อาหาร')} />
          <IconCircle icon={Landmark} bg="#E4EDEA" color="var(--pine)" label="เที่ยว" onClick={() => onQuickAction('สถานที่เที่ยว')} />
          <IconCircle icon={Sparkles} bg="var(--amber-soft)" color="var(--amber-ink)" label="AI แพลน" onClick={onAiTap} />
        </div>

        <button onClick={onAiTap} style={{ textAlign: 'left', border: 'none', cursor: 'pointer', borderRadius: 20, padding: 20, background: 'linear-gradient(135deg, var(--pine) 0%, #1C5C57 100%)', color: '#fff' }}>
          <Sparkles size={20} color="var(--amber)" style={{ marginBottom: 10 }} />
          <p className="font-display" style={{ fontSize: 17, fontWeight: 600, marginBottom: 4 }}>ให้ AI จัดทริปให้คุณ</p>
          <p style={{ fontSize: 12.5, opacity: 0.85, maxWidth: 220, lineHeight: 1.5 }}>บอกจุดหมาย งบประมาณ และสไตล์การเที่ยว ที่เหลือให้ AI จัดการ</p>
          <span style={{ display: 'inline-block', marginTop: 14, fontSize: 12, fontWeight: 700, background: 'var(--amber)', color: 'var(--on-amber-solid)', padding: '8px 14px', borderRadius: 999 }}>เริ่มวางแผนทริป</span>
        </button>

        <div>
          <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginBottom: 12 }}>
            <h2 className="font-display" style={{ fontSize: 17, fontWeight: 600, color: 'var(--ink)' }}>จุดหมายยอดนิยม</h2>
            <span style={{ fontSize: 12, color: 'var(--pine)', fontWeight: 600 }}>ดูทั้งหมด</span>
          </div>
          <div className="scrollbar-hide" style={{ display: 'flex', gap: 12, overflowX: 'auto', paddingBottom: 4 }}>
            {DESTINATIONS.map((d) => (
              <button key={d.id} onClick={() => onDestinationTap(d)} style={{ flex: '0 0 auto', width: 150, height: 180, borderRadius: 20, border: 'none', cursor: 'pointer', padding: 14, display: 'flex', flexDirection: 'column', justifyContent: 'flex-end', background: d.gradient, position: 'relative', overflow: 'hidden' }}>
                <span style={{ position: 'absolute', top: 12, right: 12 }}><Heart size={16} color="#fff" /></span>
                <p className="font-display" style={{ color: '#fff', fontSize: 17, fontWeight: 600, textAlign: 'left' }}>{d.name}</p>
                <p style={{ color: 'rgba(255,255,255,0.85)', fontSize: 10.5, textAlign: 'left', marginTop: 2 }}>{d.sub}</p>
              </button>
            ))}
          </div>
        </div>

        <div style={{ borderRadius: 20, background: 'var(--amber-soft)', padding: 16, display: 'flex', alignItems: 'center', gap: 14 }}>
          <div style={{ width: 44, height: 44, borderRadius: 14, background: 'var(--amber)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
            <Ticket size={20} color="var(--on-amber-solid)" />
          </div>
          <div>
            <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--on-amber)' }}>ลด 20% ทริปแรกของคุณ</p>
            <p style={{ fontSize: 11.5, color: 'var(--amber-ink)' }}>ใช้โค้ด SABAI20 ก่อนสิ้นเดือนนี้</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function SearchScreen({ query, setQuery, category, setCategory, results, onSelect }) {
  const categories = ['ทั้งหมด', 'ที่พัก', 'อาหาร', 'เที่ยวบิน', 'สถานที่เที่ยว'];
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div style={{ padding: '18px 20px 14px', display: 'flex', flexDirection: 'column', gap: 14, borderBottom: '1px solid var(--border)', flexShrink: 0 }}>
        <h1 className="font-display" style={{ fontSize: 20, fontWeight: 600 }}>ค้นหา</h1>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '12px 14px', borderRadius: 14, border: '1.5px solid var(--border)' }}>
          <Search size={17} color="var(--ink-soft)" />
          <input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="ค้นหาที่พัก ร้านอาหาร เที่ยวบิน..."
            style={{ border: 'none', outline: 'none', flex: 1, fontSize: 13.5, fontFamily: "'Plus Jakarta Sans', sans-serif" }} />
          <SlidersHorizontal size={17} color="var(--ink-soft)" />
        </div>
        <div className="scrollbar-hide" style={{ display: 'flex', gap: 8, overflowX: 'auto' }}>
          {categories.map((c) => <Chip key={c} label={c} active={category === c} onClick={() => setCategory(c)} />)}
        </div>
      </div>
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 12 }}>
        {results.length === 0 && <p style={{ textAlign: 'center', color: 'var(--ink-soft)', fontSize: 13, marginTop: 40 }}>ไม่พบผลลัพธ์ที่ตรงกัน ลองค้นหาด้วยคำอื่น</p>}
        {results.map((item) => {
          const Icon = TYPE_ICON[item.type];
          return (
            <button key={item.id} onClick={() => onSelect(item)} style={{ display: 'flex', gap: 14, textAlign: 'left', border: '1px solid var(--border)', borderRadius: 18, padding: 12, background: 'var(--surface)', cursor: 'pointer' }}>
              <div style={{ width: 64, height: 64, borderRadius: 14, flexShrink: 0, background: TYPE_GRADIENT[item.type], display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Icon size={24} color="#fff" />
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <p style={{ fontSize: 14, fontWeight: 700, color: 'var(--ink)' }}>{item.name}</p>
                <p style={{ fontSize: 11.5, color: 'var(--ink-soft)', margin: '2px 0 6px', display: 'flex', alignItems: 'center', gap: 4 }}><MapPin size={11} /> {item.location}</p>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <RatingStars value={item.rating} />
                  <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--pine)' }}>฿{item.price.toLocaleString()}</span>
                </div>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

function ListingDetailScreen({ item, onBack, onBook }) {
  if (!item) return null;
  const Icon = TYPE_ICON[item.type];
  const amenities = item.type === 'hotel'
    ? [{ icon: Wifi, label: 'ไวไฟฟรี' }, { icon: Waves, label: 'สระว่ายน้ำ' }, { icon: Car, label: 'ที่จอดรถ' }]
    : item.type === 'restaurant'
    ? [{ icon: UtensilsCrossed, label: 'อาหารพื้นเมือง' }, { icon: Coffee, label: 'เครื่องดื่ม' }]
    : [{ icon: Clock, label: 'เปิด 08:00–18:00' }];
  const desc = {
    hotel: 'ที่พักสไตล์ล้านนา ท่ามกลางบรรยากาศร่มรื่น ห่างจากย่านใจกลางเมืองเพียงไม่กี่นาที เหมาะสำหรับทริปพักผ่อนหรือทำงานไปด้วยเที่ยวไปด้วย',
    restaurant: 'ร้านอาหารท้องถิ่นที่คนในพื้นที่แนะนำ เสิร์ฟเมนูขึ้นชื่อด้วยวัตถุดิบสดใหม่ทุกวัน บรรยากาศเป็นกันเอง',
    attraction: 'สถานที่ท่องเที่ยวสำคัญที่ไม่ควรพลาด เหมาะกับการถ่ายภาพและเรียนรู้ประวัติศาสตร์ท้องถิ่น',
    flight: 'เที่ยวบินตรง สะดวก ตรงเวลา พร้อมบริการเช็คอินออนไลน์และเลือกที่นั่งล่วงหน้า',
  }[item.type];
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title={TYPE_LABEL[item.type]} onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
        <div style={{ height: 190, background: TYPE_GRADIENT[item.type], display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Icon size={56} color="rgba(255,255,255,0.9)" />
        </div>
        <div style={{ padding: 20 }}>
          <p className="font-display" style={{ fontSize: 21, fontWeight: 600, color: 'var(--ink)' }}>{item.name}</p>
          <p style={{ fontSize: 12.5, color: 'var(--ink-soft)', display: 'flex', alignItems: 'center', gap: 4, margin: '6px 0 10px' }}><MapPin size={12} /> {item.location}</p>
          <div style={{ display: 'flex', alignItems: 'center', gap: 14, marginBottom: 16 }}>
            <RatingStars value={item.rating} />
            <span style={{ fontSize: 12, color: 'var(--ink-soft)' }}>{item.reviews.toLocaleString()} รีวิว</span>
          </div>
          <p style={{ fontSize: 13.5, lineHeight: 1.7, color: 'var(--ink)', marginBottom: 18 }}>{desc}</p>
          <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
            {amenities.map((a, i) => {
              const AIcon = a.icon;
              return (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '8px 12px', borderRadius: 999, background: 'var(--pine-pale)' }}>
                  <AIcon size={14} color="var(--pine)" />
                  <span style={{ fontSize: 11.5, fontWeight: 600, color: 'var(--pine)' }}>{a.label}</span>
                </div>
              );
            })}
          </div>
        </div>
      </div>
      <div style={{ padding: '16px 20px', borderTop: '1px solid var(--border)', display: 'flex', alignItems: 'center', gap: 14, flexShrink: 0 }}>
        <div style={{ flex: 1 }}>
          <p style={{ fontSize: 11, color: 'var(--ink-soft)' }}>เริ่มต้น</p>
          <p className="font-display" style={{ fontSize: 19, fontWeight: 700, color: 'var(--pine)' }}>฿{item.price.toLocaleString()}</p>
        </div>
        <div style={{ flex: 1 }}><PrimaryButton onClick={onBook}>จองเลย</PrimaryButton></div>
      </div>
    </div>
  );
}

/* ---------------------------------------------------------------- */
/*  GOOGLE MAPS NAVIGATION                                            */
/* ---------------------------------------------------------------- */

const TRAVEL_MODES = [
  { id: 'driving', label: 'ขับรถ' },
  { id: 'two-wheeler', label: 'มอเตอร์ไซค์' },
  { id: 'transit', label: 'ขนส่งสาธารณะ' },
  { id: 'walking', label: 'เดิน' },
];
const MAPS_BASE = 'https://www.google.com/maps';
const placeQuery = (place, destination) => (destination && !place.includes(destination) ? `${place} ${destination}` : place);
const mapsSearchUrl = (query) => `${MAPS_BASE}/search/?api=1&query=${encodeURIComponent(query)}`;
// ปลายทางเดียว: ถ้าไม่ระบุต้นทาง Google Maps จะใช้ตำแหน่งปัจจุบันของผู้ใช้
const mapsDirUrl = ({ origin, destination, waypoints = [], mode = 'driving' }) => {
  const params = [`api=1`, `destination=${encodeURIComponent(destination)}`, `travelmode=${mode}`];
  if (origin) params.push(`origin=${encodeURIComponent(origin)}`);
  if (waypoints.length) params.push(`waypoints=${waypoints.map(encodeURIComponent).join('%7C')}`);
  return `${MAPS_BASE}/dir/?${params.join('&')}`;
};
const mapsEmbedUrl = (query) => `${MAPS_BASE}?q=${encodeURIComponent(query)}&output=embed&hl=th`;
// Google Maps รองรับจุดแวะสูงสุด 9 จุดต่อเส้นทาง
const MAX_WAYPOINTS = 9;

function MapLink({ href, icon: Icon, label, ariaLabel, solid }) {
  return (
    <a href={href} target="_blank" rel="noopener noreferrer" aria-label={ariaLabel || label}
      style={{ display: 'inline-flex', alignItems: 'center', gap: 5, padding: '6px 12px', borderRadius: 999, fontSize: 11.5, fontWeight: 700, textDecoration: 'none', border: '1.5px solid var(--pine)', background: solid ? 'var(--pine)' : 'transparent', color: solid ? '#fff' : 'var(--pine)' }}>
      <Icon size={13} /> {label}
    </a>
  );
}

function MapPreview({ query, onClose }) {
  return (
    <div style={{ border: '1px solid var(--border)', borderRadius: 16, overflow: 'hidden', background: 'var(--surface)' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '10px 12px' }}>
        <MapPin size={15} color="var(--pine)" />
        <span style={{ flex: 1, minWidth: 0, fontSize: 12.5, fontWeight: 700, color: 'var(--ink)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{query}</span>
        <button onClick={onClose} aria-label="ปิดแผนที่" style={{ background: 'none', border: 'none', cursor: 'pointer', display: 'flex', padding: 4 }}><X size={16} color="var(--ink-soft)" /></button>
      </div>
      <iframe title={`แผนที่ ${query}`} src={mapsEmbedUrl(query)} loading="lazy" referrerPolicy="no-referrer-when-downgrade" style={{ width: '100%', height: 200, border: 0, display: 'block', background: 'var(--muted-bg)' }} />
      <div style={{ padding: '10px 12px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8 }}>
        <span style={{ fontSize: 11, color: 'var(--ink-soft)' }}>หากแผนที่ไม่แสดง ให้เปิดใน Google Maps</span>
        <MapLink href={mapsSearchUrl(query)} icon={MapIcon} label="เปิดแอป" ariaLabel={`เปิด ${query} ใน Google Maps`} />
      </div>
    </div>
  );
}

function TripItinerary({ itinerary, destination }) {
  const [mode, setMode] = useState('driving');
  const [focus, setFocus] = useState(null);
  const hasPlaces = itinerary.days.some((d) => d.items.some((it) => it.place));
  return (
    <>
      {hasPlaces && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Navigation size={16} color="var(--pine)" />
            <p className="font-display" style={{ fontSize: 15, fontWeight: 600 }}>นำทางด้วย Google Maps</p>
          </div>
          <p style={{ fontSize: 11.5, color: 'var(--ink-soft)', lineHeight: 1.6 }}>เลือกวิธีเดินทาง แล้วกด "นำทาง" ที่แต่ละจุด หรือดูเส้นทางทั้งวัน ระบบจะเปิด Google Maps ในแท็บใหม่</p>
          <div role="radiogroup" aria-label="วิธีเดินทาง" className="scrollbar-hide" style={{ display: 'flex', gap: 8, overflowX: 'auto' }}>
            {TRAVEL_MODES.map((m) => (
              <button key={m.id} role="radio" aria-checked={mode === m.id} onClick={() => setMode(m.id)}
                style={{ padding: '7px 14px', borderRadius: 999, fontSize: 12, fontWeight: 600, whiteSpace: 'nowrap', cursor: 'pointer', border: mode === m.id ? '1.5px solid var(--pine)' : '1.5px solid var(--border)', background: mode === m.id ? 'var(--pine)' : 'var(--surface)', color: mode === m.id ? '#fff' : 'var(--ink)' }}>
                {m.label}
              </button>
            ))}
          </div>
          {focus && <MapPreview query={focus} onClose={() => setFocus(null)} />}
        </div>
      )}
      {itinerary.days.map((day, i) => {
        const stops = day.items.filter((it) => it.place).map((it) => placeQuery(it.place, destination));
        const canRoute = stops.length >= 2;
        const routeUrl = canRoute ? mapsDirUrl({ origin: stops[0], destination: stops[stops.length - 1], waypoints: stops.slice(1, -1).slice(0, MAX_WAYPOINTS), mode }) : null;
        return (
          <div key={i}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8, marginBottom: 10 }}>
              <p style={{ fontSize: 13.5, fontWeight: 700, color: 'var(--pine)' }}>{day.title}</p>
              {canRoute && <MapLink href={routeUrl} icon={RouteIcon} label={`เส้นทางทั้งวัน (${stops.length} จุด)`} ariaLabel={`ดูเส้นทางทั้งวัน ${day.title}`} />}
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {day.items.map((it, j) => {
                const query = it.place ? placeQuery(it.place, destination) : null;
                // จุดแรกของทริปใช้ตำแหน่งปัจจุบันเป็นต้นทาง จุดถัดไปเริ่มจากจุดก่อนหน้าในวันเดียวกัน
                const prev = day.items.slice(0, j).reverse().find((x) => x.place);
                const navUrl = query ? mapsDirUrl({ origin: prev ? placeQuery(prev.place, destination) : undefined, destination: query, mode }) : null;
                return (
                  <div key={j} style={{ display: 'flex', gap: 12 }}>
                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                      <div style={{ width: 8, height: 8, borderRadius: 999, background: 'var(--amber)', marginTop: 5 }} />
                      {j < day.items.length - 1 && <div style={{ width: 1.5, flex: 1, background: 'var(--border)', marginTop: 2 }} />}
                    </div>
                    <div style={{ paddingBottom: 10, flex: 1, minWidth: 0 }}>
                      <p style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--ink-soft)' }}>{it.time}</p>
                      <p style={{ fontSize: 13, color: 'var(--ink)', marginTop: 1 }}>{it.text}</p>
                      {query && (
                        <div style={{ display: 'flex', gap: 8, marginTop: 8, flexWrap: 'wrap' }}>
                          <MapLink href={navUrl} icon={Navigation} label="นำทาง" ariaLabel={`นำทางไป ${it.place}`} solid />
                          <button onClick={() => setFocus(query)} aria-label={`ดูแผนที่ ${it.place}`}
                            style={{ display: 'inline-flex', alignItems: 'center', gap: 5, padding: '6px 12px', borderRadius: 999, fontSize: 11.5, fontWeight: 700, cursor: 'pointer', border: '1.5px solid var(--border)', background: 'transparent', color: 'var(--ink)' }}>
                            <MapIcon size={13} /> ดูแผนที่
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        );
      })}
    </>
  );
}

function AiPlannerScreen({ destination, setDestination, days, setDays, budget, setBudget, styles, toggleStyle, travelers, setTravelers, generating, result, onGenerate, onReset, onSave, onBookTrip, saved }) {
  if (generating) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', gap: 18, padding: 20, textAlign: 'center' }}>
        <div style={{ width: 64, height: 64, borderRadius: 999, background: 'var(--pine-pale)', display: 'flex', alignItems: 'center', justifyContent: 'center', animation: 'spin-slow 2.2s linear infinite' }}>
          <Sparkles size={28} color="var(--pine)" />
        </div>
        <p className="font-display" style={{ fontSize: 17, fontWeight: 600 }}>AI กำลังจัดทริปให้คุณ...</p>
        <p style={{ fontSize: 12.5, color: 'var(--ink-soft)', maxWidth: 230 }}>กำลังวิเคราะห์งบประมาณ ระยะทาง และสไตล์การเที่ยวที่คุณเลือก</p>
      </div>
    );
  }

  if (result) {
    const total = result.budgetBreakdown.reduce((s, b) => s + b.value, 0);
    const maxVal = Math.max(...result.budgetBreakdown.map((b) => b.value));
    return (
      <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <div style={{ padding: '18px 20px 12px', borderBottom: '1px solid var(--border)', flexShrink: 0 }}>
          <button onClick={onReset} style={{ background: 'none', border: 'none', color: 'var(--pine)', fontSize: 12, fontWeight: 700, marginBottom: 8, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}><ArrowLeft size={14} /> วางแผนใหม่</button>
          <h1 className="font-display" style={{ fontSize: 19, fontWeight: 600 }}>ทริป {destination} · {days} วัน</h1>
          <p style={{ fontSize: 12, color: 'var(--ink-soft)' }}>{travelers} ผู้เดินทาง · จัดโดย AI</p>
        </div>
        <div className="scrollbar-hide fade-up" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: '18px 20px', display: 'flex', flexDirection: 'column', gap: 20 }}>
          <TripItinerary itinerary={result} destination={destination} />

          <div>
            <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 10 }}>ประมาณการค่าใช้จ่าย</p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 10 }}>
              {result.budgetBreakdown.map((b, i) => (
                <div key={i}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 4 }}>
                    <span style={{ color: 'var(--ink)' }}>{b.label}</span>
                    <span style={{ fontWeight: 700, color: 'var(--ink)' }}>฿{b.value.toLocaleString()}</span>
                  </div>
                  <div style={{ height: 6, borderRadius: 999, background: 'var(--pine-pale)' }}>
                    <div style={{ height: '100%', width: `${(b.value / maxVal) * 100}%`, borderRadius: 999, background: 'var(--amber)' }} />
                  </div>
                </div>
              ))}
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', paddingTop: 10, borderTop: '1px dashed var(--border)' }}>
              <span style={{ fontSize: 13.5, fontWeight: 700 }}>รวมทั้งหมด</span>
              <span className="font-display" style={{ fontSize: 17, fontWeight: 700, color: 'var(--pine)' }}>฿{total.toLocaleString()}</span>
            </div>
          </div>
        </div>
        <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', display: 'flex', gap: 10, flexShrink: 0 }}>
          <div style={{ flex: 1 }}><SecondaryButton onClick={onSave}>{saved ? 'บันทึกแล้ว ✓' : 'บันทึกทริป'}</SecondaryButton></div>
          <div style={{ flex: 1.4 }}><PrimaryButton onClick={onBookTrip}>จองทริปนี้</PrimaryButton></div>
        </div>
      </div>
    );
  }

  return (
    <div className="scrollbar-hide" style={{ height: '100%', overflowY: 'auto', padding: '20px 20px 28px', display: 'flex', flexDirection: 'column', gap: 22 }}>
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
          <Sparkles size={18} color="var(--amber)" />
          <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--pine)' }}>ผู้ช่วยวางแผนทริป</span>
        </div>
        <h1 className="font-display" style={{ fontSize: 22, fontWeight: 600 }}>ให้ AI จัดทริปในฝันของคุณ</h1>
        <p style={{ fontSize: 12.5, color: 'var(--ink-soft)', marginTop: 4 }}>ตอบคำถามสั้นๆ แล้วปล่อยให้ AI จัดการที่เหลือ</p>
      </div>

      <div>
        <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>อยากไปที่ไหน?</p>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
          {DESTINATIONS.map((d) => <Chip key={d.id} label={d.name} active={destination === d.name} onClick={() => setDestination(d.name)} />)}
        </div>
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700 }}>กี่วัน?</p>
          <p style={{ fontSize: 11.5, color: 'var(--ink-soft)' }}>รวมวันเดินทาง</p>
        </div>
        <Stepper value={days} onChange={setDays} min={1} max={14} />
      </div>

      <div>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
          <p style={{ fontSize: 13, fontWeight: 700 }}>งบประมาณโดยประมาณ</p>
          <span className="font-display" style={{ fontSize: 15, fontWeight: 700, color: 'var(--pine)' }}>฿{budget.toLocaleString()}</span>
        </div>
        <input type="range" min={3000} max={50000} step={500} value={budget} onChange={(e) => setBudget(Number(e.target.value))} style={{ width: '100%', accentColor: '#10403D' }} />
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 10.5, color: 'var(--ink-soft)' }}>
          <span>฿3,000</span><span>฿50,000</span>
        </div>
      </div>

      <div>
        <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>สไตล์การเที่ยว</p>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
          {STYLE_OPTIONS.map((s) => <Chip key={s} label={s} active={styles.includes(s)} onClick={() => toggleStyle(s)} />)}
        </div>
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <p style={{ fontSize: 13, fontWeight: 700 }}>จำนวนผู้เดินทาง</p>
        <Stepper value={travelers} onChange={setTravelers} min={1} max={10} />
      </div>

      <PrimaryButton onClick={onGenerate}>ให้ AI จัดแผนให้ ✨</PrimaryButton>
    </div>
  );
}

function TripsScreen({ savedTrips, onStartAi, onSelectTrip }) {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <h1 className="font-display" style={{ fontSize: 20, fontWeight: 600 }}>ทริปของฉัน</h1>
        {savedTrips.length === 0 ? (
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 14, textAlign: 'center', padding: 20 }}>
            <div style={{ width: 64, height: 64, borderRadius: 999, background: 'var(--pine-pale)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Briefcase size={26} color="var(--pine)" />
            </div>
            <p style={{ fontSize: 14, fontWeight: 700, color: 'var(--ink)' }}>คุณยังไม่มีทริปที่บันทึกไว้</p>
            <p style={{ fontSize: 12, color: 'var(--ink-soft)', maxWidth: 220 }}>เริ่มวางแผนทริปแรกของคุณด้วย AI ได้เลยตอนนี้</p>
            <div style={{ width: 180 }}><PrimaryButton onClick={onStartAi}>เริ่มวางแผน</PrimaryButton></div>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {savedTrips.map((t) => (
              <button key={t.id} onClick={() => onSelectTrip(t)} style={{ textAlign: 'left', border: '1px solid var(--border)', borderRadius: 18, padding: 16, display: 'flex', gap: 14, alignItems: 'center', background: 'var(--surface)', cursor: 'pointer' }}>
                <div style={{ width: 52, height: 52, borderRadius: 14, background: 'linear-gradient(135deg,#1C5C57,#89A66B)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                  <MapPin size={22} color="#fff" />
                </div>
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 14, fontWeight: 700 }}>{t.destination}</p>
                  <p style={{ fontSize: 11.5, color: 'var(--ink-soft)' }}>{t.days} วัน · งบ ฿{t.budget.toLocaleString()}</p>
                </div>
                <ChevronRight size={18} color="var(--ink-soft)" />
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function TripDetailScreen({ trip, onBack }) {
  if (!trip) return null;
  const total = trip.itinerary.budgetBreakdown.reduce((s, b) => s + b.value, 0);
  const maxVal = Math.max(...trip.itinerary.budgetBreakdown.map((b) => b.value));
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="รายละเอียดทริป" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
        <div style={{ height: 130, background: 'linear-gradient(135deg,#1C5C57,#89A66B)', display: 'flex', flexDirection: 'column', justifyContent: 'flex-end', padding: 18 }}>
          <p className="font-display" style={{ color: '#fff', fontSize: 22, fontWeight: 600 }}>{trip.destination}</p>
          <p style={{ color: 'rgba(255,255,255,0.85)', fontSize: 12 }}>{trip.days} วัน · {trip.travelers} ผู้เดินทาง</p>
        </div>
        <div style={{ padding: '18px 20px', display: 'flex', flexDirection: 'column', gap: 20 }}>
          <TripItinerary itinerary={trip.itinerary} destination={trip.destination} />

          <div>
            <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 10 }}>ประมาณการค่าใช้จ่าย</p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 10 }}>
              {trip.itinerary.budgetBreakdown.map((b, i) => (
                <div key={i}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 4 }}>
                    <span style={{ color: 'var(--ink)' }}>{b.label}</span>
                    <span style={{ fontWeight: 700, color: 'var(--ink)' }}>฿{b.value.toLocaleString()}</span>
                  </div>
                  <div style={{ height: 6, borderRadius: 999, background: 'var(--pine-pale)' }}>
                    <div style={{ height: '100%', width: `${(b.value / maxVal) * 100}%`, borderRadius: 999, background: 'var(--amber)' }} />
                  </div>
                </div>
              ))}
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', paddingTop: 10, borderTop: '1px dashed var(--border)' }}>
              <span style={{ fontSize: 13.5, fontWeight: 700 }}>รวมทั้งหมด</span>
              <span className="font-display" style={{ fontSize: 17, fontWeight: 700, color: 'var(--pine)' }}>฿{total.toLocaleString()}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function ProfileScreen({ onLogout, onWalletTap, onReviewsTap, onCouponTap, onHelpTap, onBookingsTap, onSettingsTap, onCommunityTap }) {
  const items = [
    { icon: Users, label: 'ชุมชนของฉัน', onClick: onCommunityTap },
    { icon: FileText, label: 'การจองของฉัน', onClick: onBookingsTap },
    { icon: Wallet, label: 'กระเป๋าเงิน', onClick: onWalletTap },
    { icon: Ticket, label: 'คูปองของฉัน', onClick: onCouponTap },
    { icon: Star, label: 'รีวิวของฉัน', onClick: onReviewsTap },
    { icon: Settings, label: 'การตั้งค่า', onClick: onSettingsTap },
    { icon: HelpCircle, label: 'ช่วยเหลือ', onClick: onHelpTap },
  ];
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <div style={{ width: 60, height: 60, borderRadius: 999, background: 'var(--amber-soft)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20, fontWeight: 700, color: 'var(--pine)' }}>น</div>
          <div>
            <p className="font-display" style={{ fontSize: 17, fontWeight: 600 }}>หนาน ใจดี</p>
            <p style={{ fontSize: 12, color: 'var(--ink-soft)' }}>nan.jaidee@email.com</p>
          </div>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', borderRadius: 18, border: '1px solid var(--border)', overflow: 'hidden' }}>
          {items.map((it, i) => {
            const Icon = it.icon;
            return (
              <button key={i} onClick={it.onClick} style={{ width: '100%', textAlign: 'left', background: 'none', borderTop: 'none', borderLeft: 'none', borderRight: 'none', display: 'flex', alignItems: 'center', gap: 12, padding: '14px 16px', borderBottom: i < items.length - 1 ? '1px solid var(--border)' : 'none', cursor: 'pointer' }}>
                <Icon size={17} color="var(--pine)" />
                <span style={{ fontSize: 13.5, flex: 1, color: 'var(--ink)' }}>{it.label}</span>
                <ChevronRight size={16} color="var(--ink-soft)" />
              </button>
            );
          })}
        </div>
        <button onClick={onLogout} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8, padding: 14, borderRadius: 16, border: '1.5px solid var(--danger)', background: 'transparent', color: 'var(--danger)', fontWeight: 700, fontSize: 13.5, cursor: 'pointer' }}>
          <LogOut size={16} /> ออกจากระบบ
        </button>
      </div>
    </div>
  );
}

function WalletScreen({ balance, transactions, onBack, onTopUp }) {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="กระเป๋าเงิน" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ borderRadius: 20, padding: 20, background: 'linear-gradient(135deg, var(--pine) 0%, #1C5C57 100%)', color: '#fff' }}>
          <p style={{ fontSize: 12, opacity: 0.8 }}>ยอดคงเหลือ</p>
          <p className="font-display" style={{ fontSize: 30, fontWeight: 600, margin: '6px 0 16px' }}>฿{balance.toLocaleString()}</p>
          <div style={{ display: 'flex', gap: 10 }}>
            <button onClick={onTopUp} style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6, padding: 10, borderRadius: 12, border: 'none', background: 'var(--amber)', color: 'var(--on-amber-solid)', fontWeight: 700, fontSize: 12.5, cursor: 'pointer' }}>
              <ArrowDown size={14} /> เติมเงิน
            </button>
            <button style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6, padding: 10, borderRadius: 12, border: '1.5px solid rgba(255,255,255,0.4)', background: 'transparent', color: '#fff', fontWeight: 700, fontSize: 12.5, cursor: 'pointer' }}>
              <ArrowUp size={14} /> ถอนเงิน
            </button>
          </div>
        </div>
        <div>
          <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 12 }}>ธุรกรรมล่าสุด</p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {transactions.map((t) => (
              <div key={t.id} style={{ display: 'flex', alignItems: 'center', gap: 12, border: '1px solid var(--border)', borderRadius: 16, padding: 12 }}>
                <div style={{ width: 38, height: 38, borderRadius: 999, background: t.type === 'in' ? 'var(--pine-pale)' : 'var(--danger-soft)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                  {t.type === 'in' ? <ArrowDown size={16} color="var(--pine)" /> : <ArrowUp size={16} color="var(--danger)" />}
                </div>
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13, fontWeight: 600, color: 'var(--ink)' }}>{t.label}</p>
                  <p style={{ fontSize: 11, color: 'var(--ink-soft)' }}>{t.date}</p>
                </div>
                <span style={{ fontSize: 13, fontWeight: 700, color: t.type === 'in' ? 'var(--pine)' : 'var(--danger)' }}>{t.type === 'in' ? '+' : '-'}฿{t.amount.toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function BookingScreen({ item, onBack, onPay }) {
  if (!item) return null;
  const fee = Math.round(item.price * 0.05);
  const total = item.price + fee;
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="สรุปการจอง" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 18 }}>
        <div style={{ border: '1px solid var(--border)', borderRadius: 18, padding: 16 }}>
          <p style={{ fontSize: 11, fontWeight: 700, color: 'var(--ink-soft)', marginBottom: 6 }}>{TYPE_LABEL[item.type] || 'รายการ'}</p>
          <p className="font-display" style={{ fontSize: 16, fontWeight: 600 }}>{item.title}</p>
          <p style={{ fontSize: 12, color: 'var(--ink-soft)', marginTop: 2 }}>{item.subtitle}</p>
        </div>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>รหัสส่วนลด</p>
          <div style={{ display: 'flex', gap: 8 }}>
            <div style={{ flex: 1 }}><TextField icon={Ticket} placeholder="กรอกโค้ดส่วนลด" value="" onChange={() => {}} disabled /></div>
            <button style={{ padding: '0 18px', borderRadius: 14, border: '1.5px solid var(--pine)', background: 'transparent', color: 'var(--pine)', fontWeight: 700, fontSize: 13 }}>ใช้</button>
          </div>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}><span style={{ color: 'var(--ink-soft)' }}>ราคารวม</span><span style={{ fontWeight: 600 }}>฿{item.price.toLocaleString()}</span></div>
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}><span style={{ color: 'var(--ink-soft)' }}>ค่าธรรมเนียมบริการ</span><span style={{ fontWeight: 600 }}>฿{fee.toLocaleString()}</span></div>
          <div style={{ display: 'flex', justifyContent: 'space-between', paddingTop: 10, borderTop: '1px dashed var(--border)' }}>
            <span style={{ fontWeight: 700 }}>ยอดชำระทั้งหมด</span>
            <span className="font-display" style={{ fontSize: 18, fontWeight: 700, color: 'var(--pine)' }}>฿{total.toLocaleString()}</span>
          </div>
        </div>
      </div>
      <div style={{ padding: '16px 20px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        <PrimaryButton onClick={onPay}>ดำเนินการชำระเงิน</PrimaryButton>
      </div>
    </div>
  );
}

function PaymentScreen({ total, method, setMethod, onBack, onConfirm, processing, walletBalance, error }) {
  const methods = [
    { id: 'promptpay', label: 'พร้อมเพย์ (PromptPay)', icon: QrCode, note: 'สแกน QR เพื่อชำระเงินทันที' },
    { id: 'card', label: 'บัตรเครดิต/เดบิต', icon: CreditCard, note: 'Visa, Mastercard, JCB' },
    { id: 'wallet', label: 'Sabai Wallet', icon: Wallet, note: `คงเหลือ ฿${walletBalance.toLocaleString()}` },
  ];
  if (processing) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', gap: 16 }}>
        <div style={{ width: 56, height: 56, borderRadius: 999, border: '4px solid var(--pine-pale)', borderTopColor: 'var(--pine)', animation: 'spin-slow 0.9s linear infinite' }} />
        <p style={{ fontSize: 13.5, fontWeight: 700 }}>กำลังดำเนินการชำระเงิน...</p>
        <p style={{ fontSize: 11.5, color: 'var(--ink-soft)' }}>กรุณาอย่าปิดหน้าจอ</p>
      </div>
    );
  }
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="ชำระเงิน" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}>
        {methods.map((m) => {
          const Icon = m.icon;
          const active = method === m.id;
          return (
            <button key={m.id} onClick={() => setMethod(m.id)} style={{ display: 'flex', alignItems: 'center', gap: 14, padding: 16, borderRadius: 16, border: active ? '2px solid var(--pine)' : '1.5px solid var(--border)', background: active ? 'var(--pine-pale)' : 'var(--surface)', cursor: 'pointer', textAlign: 'left' }}>
              <Icon size={22} color="var(--pine)" />
              <div style={{ flex: 1 }}>
                <p style={{ fontSize: 13.5, fontWeight: 700, color: 'var(--ink)' }}>{m.label}</p>
                <p style={{ fontSize: 11, color: 'var(--ink-soft)' }}>{m.note}</p>
              </div>
              <div style={{ width: 20, height: 20, borderRadius: 999, border: '2px solid var(--pine)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                {active && <div style={{ width: 10, height: 10, borderRadius: 999, background: 'var(--pine)' }} />}
              </div>
            </button>
          );
        })}
        {method === 'card' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 6 }}>
            <TextField icon={CreditCard} placeholder="หมายเลขบัตร" value="" onChange={() => {}} disabled />
            <div style={{ display: 'flex', gap: 10 }}>
              <div style={{ flex: 1 }}><TextField icon={Calendar} placeholder="MM/YY" value="" onChange={() => {}} disabled /></div>
              <div style={{ flex: 1 }}><TextField icon={Lock} placeholder="CVV" value="" onChange={() => {}} disabled /></div>
            </div>
          </div>
        )}
      </div>
      <div style={{ padding: '16px 20px', borderTop: '1px solid var(--border)', display: 'flex', flexDirection: 'column', gap: 10, flexShrink: 0 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 13, color: 'var(--ink-soft)' }}>ยอดชำระ</span>
          <span className="font-display" style={{ fontSize: 17, fontWeight: 700, color: 'var(--pine)' }}>฿{total.toLocaleString()}</span>
        </div>
        {error && <p role="alert" style={{ fontSize: 12, fontWeight: 600, color: 'var(--danger)' }}>{error}</p>}
        <PrimaryButton onClick={onConfirm}>ยืนยันการชำระเงิน</PrimaryButton>
      </div>
    </div>
  );
}

function ConfirmationScreen({ bookingNumber, item, onHome, onViewBookings }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', padding: 30, textAlign: 'center', gap: 16 }}>
      <div className="fade-up" style={{ width: 76, height: 76, borderRadius: 999, background: 'var(--pine-pale)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <CheckCircle2 size={40} color="var(--pine)" />
      </div>
      <p className="font-display" style={{ fontSize: 22, fontWeight: 600 }}>จองสำเร็จ!</p>
      <p style={{ fontSize: 13, color: 'var(--ink-soft)' }}>เราได้ส่งรายละเอียดการจองไปยังอีเมลของคุณแล้ว</p>
      <div style={{ width: '100%', border: '1px dashed var(--border)', borderRadius: 16, padding: 16 }}>
        <p style={{ fontSize: 11, color: 'var(--ink-soft)' }}>หมายเลขการจอง</p>
        <p className="font-display" style={{ fontSize: 18, fontWeight: 700, color: 'var(--pine)' }}>{bookingNumber}</p>
        {item && <p style={{ fontSize: 12.5, color: 'var(--ink)', marginTop: 8 }}>{item.title}</p>}
      </div>
      <div style={{ width: '100%', marginTop: 10, display: 'flex', flexDirection: 'column', gap: 10 }}>
        <PrimaryButton onClick={onViewBookings}>ดูการจองของฉัน</PrimaryButton>
        <SecondaryButton onClick={onHome}>กลับสู่หน้าแรก</SecondaryButton>
      </div>
    </div>
  );
}

function NotificationScreen({ notifications, onBack, onMarkAll, onMarkOne }) {
  const hasUnread = notifications.some((n) => !n.read);
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="การแจ้งเตือน" onBack={onBack} right={
        hasUnread ? <button onClick={onMarkAll} style={{ background: 'none', border: 'none', color: 'var(--pine)', fontSize: 12, fontWeight: 700, cursor: 'pointer' }}>อ่านทั้งหมด</button> : null
      } />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: '12px 20px 20px', display: 'flex', flexDirection: 'column', gap: 10 }}>
        {notifications.length === 0 && <p style={{ textAlign: 'center', color: 'var(--ink-soft)', fontSize: 13, marginTop: 40 }}>ยังไม่มีการแจ้งเตือน</p>}
        {notifications.map((n) => {
          const Icon = NOTIF_ICON[n.type];
          return (
            <button key={n.id} onClick={() => onMarkOne(n.id)} style={{ display: 'flex', gap: 12, textAlign: 'left', border: '1px solid var(--border)', borderRadius: 16, padding: 14, background: n.read ? 'var(--surface)' : 'var(--pine-pale)', cursor: 'pointer' }}>
              <div style={{ width: 40, height: 40, borderRadius: 999, background: n.read ? 'var(--pine-pale)' : '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0, position: 'relative' }}>
                <Icon size={17} color="var(--pine)" />
                {!n.read && <span style={{ position: 'absolute', top: -1, right: -1, width: 9, height: 9, borderRadius: 999, background: 'var(--danger)', border: '2px solid #fff' }} />}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
                  <p style={{ fontSize: 13, fontWeight: n.read ? 600 : 700, color: 'var(--ink)' }}>{n.title}</p>
                  <span style={{ fontSize: 10.5, color: 'var(--ink-soft)', flexShrink: 0 }}>{n.time}</span>
                </div>
                <p style={{ fontSize: 12, color: 'var(--ink-soft)', marginTop: 3, lineHeight: 1.5 }}>{n.message}</p>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

function StarPicker({ value, onChange }) {
  return (
    <div style={{ display: 'flex', gap: 8 }}>
      {[1, 2, 3, 4, 5].map((n) => (
        <button key={n} onClick={() => onChange(n)} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 2 }}>
          <Star size={30} fill={n <= value ? 'var(--amber)' : 'none'} color="var(--amber)" strokeWidth={1.5} />
        </button>
      ))}
    </div>
  );
}

function TextArea({ placeholder, value, onChange }) {
  return (
    <textarea placeholder={placeholder} value={value} onChange={onChange} rows={5}
      style={{ width: '100%', border: '1.5px solid var(--border)', borderRadius: 14, padding: '12px 14px', fontSize: 13.5, fontFamily: "'Plus Jakarta Sans', sans-serif", color: 'var(--ink)', outline: 'none', resize: 'none' }} />
  );
}

function ReviewsScreen({ reviews, onBack, onWriteReview }) {
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="รีวิวของฉัน" onBack={onBack} right={
        <button onClick={onWriteReview} style={{ display: 'flex', alignItems: 'center', gap: 4, background: 'none', border: 'none', color: 'var(--pine)', fontSize: 12, fontWeight: 700, cursor: 'pointer' }}>
          <Plus size={14} /> เขียนรีวิว
        </button>
      } />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}>
        {reviews.length === 0 && (
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 14, textAlign: 'center', padding: 20 }}>
            <div style={{ width: 64, height: 64, borderRadius: 999, background: 'var(--pine-pale)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Star size={26} color="var(--pine)" />
            </div>
            <p style={{ fontSize: 14, fontWeight: 700 }}>คุณยังไม่มีรีวิว</p>
            <p style={{ fontSize: 12, color: 'var(--ink-soft)', maxWidth: 220 }}>แบ่งปันประสบการณ์ของคุณให้นักเดินทางคนอื่น</p>
            <div style={{ width: 160 }}><PrimaryButton onClick={onWriteReview}>เขียนรีวิวแรก</PrimaryButton></div>
          </div>
        )}
        {reviews.map((r) => {
          const Icon = TYPE_ICON[r.itemType];
          return (
            <div key={r.id} style={{ border: '1px solid var(--border)', borderRadius: 18, padding: 16 }}>
              <div style={{ display: 'flex', gap: 12, marginBottom: 10 }}>
                <div style={{ width: 44, height: 44, borderRadius: 12, background: TYPE_GRADIENT[r.itemType], display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                  <Icon size={18} color="#fff" />
                </div>
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: 13.5, fontWeight: 700, color: 'var(--ink)' }}>{r.itemName}</p>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 3 }}>
                    <div style={{ display: 'flex', gap: 1 }}>
                      {[1, 2, 3, 4, 5].map((n) => <Star key={n} size={12} fill={n <= r.rating ? 'var(--amber)' : 'none'} color="var(--amber)" />)}
                    </div>
                    <span style={{ fontSize: 10.5, color: 'var(--ink-soft)' }}>{r.date}</span>
                  </div>
                </div>
              </div>
              <p style={{ fontSize: 12.5, color: 'var(--ink)', lineHeight: 1.6 }}>{r.comment}</p>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function WriteReviewScreen({ selectedItem, onSelectItem, rating, setRating, comment, setComment, onBack, onSubmit }) {
  const canSubmit = selectedItem && rating > 0 && comment.trim().length > 0;
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="เขียนรีวิว" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>คุณกำลังรีวิวอะไร?</p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {LISTINGS.map((item) => {
              const Icon = TYPE_ICON[item.type];
              const active = selectedItem && selectedItem.id === item.id;
              return (
                <button key={item.id} onClick={() => onSelectItem(item)} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: 12, borderRadius: 14, border: active ? '2px solid var(--pine)' : '1.5px solid var(--border)', background: active ? 'var(--pine-pale)' : 'var(--surface)', cursor: 'pointer', textAlign: 'left' }}>
                  <div style={{ width: 36, height: 36, borderRadius: 10, background: TYPE_GRADIENT[item.type], display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    <Icon size={16} color="#fff" />
                  </div>
                  <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--ink)' }}>{item.name}</span>
                </button>
              );
            })}
          </div>
        </div>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>ให้คะแนนกี่ดาว?</p>
          <StarPicker value={rating} onChange={setRating} />
        </div>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>เล่าประสบการณ์ของคุณ</p>
          <TextArea placeholder="บอกเล่าประสบการณ์ให้นักเดินทางคนอื่นฟังหน่อย..." value={comment} onChange={(e) => setComment(e.target.value)} />
        </div>
      </div>
      <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        <PrimaryButton onClick={onSubmit} disabled={!canSubmit}>ส่งรีวิว</PrimaryButton>
      </div>
    </div>
  );
}

function CouponCard({ coupon, onCopy, copiedId }) {
  const statusMeta = {
    available: { label: 'ใช้ได้', bg: 'var(--pine-pale)', color: 'var(--pine)' },
    used: { label: 'ใช้แล้ว', bg: 'var(--muted-bg)', color: 'var(--ink-soft)' },
    expired: { label: 'หมดอายุ', bg: 'var(--danger-soft)', color: 'var(--danger)' },
  }[coupon.status];
  const isActive = coupon.status === 'available';
  return (
    <div style={{ display: 'flex', borderRadius: 18, border: '1px solid var(--border)', overflow: 'hidden', opacity: isActive ? 1 : 0.6 }}>
      <div style={{ width: 10, background: isActive ? 'var(--amber)' : 'var(--border)', flexShrink: 0 }} />
      <div style={{ flex: 1, padding: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 6 }}>
          <p className="font-display" style={{ fontSize: 15, fontWeight: 600, color: 'var(--ink)' }}>{coupon.title}</p>
          <span style={{ fontSize: 10.5, fontWeight: 700, padding: '4px 10px', borderRadius: 999, background: statusMeta.bg, color: statusMeta.color, flexShrink: 0 }}>{statusMeta.label}</span>
        </div>
        <p style={{ fontSize: 12, color: 'var(--ink-soft)', lineHeight: 1.5, marginBottom: 10 }}>{coupon.detail}</p>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ fontFamily: 'monospace', fontSize: 13, fontWeight: 700, letterSpacing: 0.5, color: 'var(--ink)', background: 'var(--paper)', padding: '4px 10px', borderRadius: 8, border: '1px dashed var(--border)' }}>{coupon.code}</span>
            {isActive && (
              <button onClick={() => onCopy(coupon)} style={{ background: 'none', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 3, color: 'var(--pine)', fontSize: 11, fontWeight: 700 }}>
                {copiedId === coupon.id ? <Check size={13} /> : <Copy size={13} />}
                {copiedId === coupon.id ? 'คัดลอกแล้ว' : 'คัดลอก'}
              </button>
            )}
          </div>
          <span style={{ fontSize: 10.5, color: 'var(--ink-soft)' }}>หมดอายุ {coupon.expiry}</span>
        </div>
      </div>
    </div>
  );
}

function CouponScreen({ coupons, onBack, onApply }) {
  const [tab, setTab] = useState('available');
  const [copiedId, setCopiedId] = useState(null);
  const [applyCode, setApplyCode] = useState('');
  const [applyMsg, setApplyMsg] = useState(null);
  const tabs = [
    { id: 'available', label: 'ใช้ได้' },
    { id: 'used', label: 'ใช้แล้ว' },
    { id: 'expired', label: 'หมดอายุ' },
  ];
  const filtered = coupons.filter((c) => c.status === tab);
  const handleCopy = (coupon) => {
    setCopiedId(coupon.id);
    setTimeout(() => setCopiedId(null), 1500);
  };
  const handleApply = () => {
    const code = applyCode.trim().toUpperCase();
    if (!code) return;
    const found = coupons.find((c) => c.code === code);
    if (!found) {
      setApplyMsg({ ok: false, text: 'ไม่พบรหัสคูปองนี้' });
    } else if (found.status === 'used') {
      setApplyMsg({ ok: false, text: 'คุณใช้คูปองนี้ไปแล้ว' });
    } else if (found.status === 'expired') {
      setApplyMsg({ ok: false, text: 'คูปองนี้หมดอายุแล้ว' });
    } else {
      onApply(found);
      setApplyMsg({ ok: true, text: `เพิ่มคูปอง "${found.title}" เรียบร้อยแล้ว` });
      setApplyCode('');
      setTab('available');
    }
    setTimeout(() => setApplyMsg(null), 2500);
  };
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="คูปองของฉัน" onBack={onBack} />
      <div style={{ padding: '14px 20px 0', flexShrink: 0, display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'flex', gap: 8 }}>
          <div style={{ flex: 1 }}><TextField icon={Ticket} placeholder="กรอกรหัสคูปองที่นี่" value={applyCode} onChange={(e) => setApplyCode(e.target.value)} /></div>
          <button onClick={handleApply} style={{ padding: '0 18px', borderRadius: 14, border: 'none', background: 'var(--pine)', color: '#fff', fontWeight: 700, fontSize: 13, cursor: 'pointer' }}>เพิ่ม</button>
        </div>
        {applyMsg && (
          <p style={{ fontSize: 12, fontWeight: 600, color: applyMsg.ok ? 'var(--pine)' : 'var(--danger)' }}>{applyMsg.text}</p>
        )}
        <div style={{ display: 'flex', gap: 8 }}>
          {tabs.map((t) => <Chip key={t.id} label={t.label} active={tab === t.id} onClick={() => setTab(t.id)} />)}
        </div>
      </div>
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: '16px 20px 20px', display: 'flex', flexDirection: 'column', gap: 12 }}>
        {filtered.length === 0 && <p style={{ textAlign: 'center', color: 'var(--ink-soft)', fontSize: 13, marginTop: 40 }}>ไม่มีคูปองในหมวดนี้</p>}
        {filtered.map((c) => <CouponCard key={c.id} coupon={c} onCopy={handleCopy} copiedId={copiedId} />)}
      </div>
    </div>
  );
}

function FaqItem({ faq, open, onToggle }) {
  return (
    <div style={{ border: '1px solid var(--border)', borderRadius: 16, overflow: 'hidden' }}>
      <button onClick={onToggle} style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10, padding: 14, background: 'none', border: 'none', cursor: 'pointer', textAlign: 'left' }}>
        <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--ink)' }}>{faq.q}</span>
        {open ? <ChevronUp size={16} color="var(--ink-soft)" style={{ flexShrink: 0 }} /> : <ChevronDown size={16} color="var(--ink-soft)" style={{ flexShrink: 0 }} />}
      </button>
      {open && <p style={{ padding: '0 14px 14px', fontSize: 12.5, color: 'var(--ink-soft)', lineHeight: 1.6 }}>{faq.a}</p>}
    </div>
  );
}

function HelpScreen({ onBack, onOpenTicket, tickets }) {
  const [openFaqId, setOpenFaqId] = useState(null);
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="ช่วยเหลือ" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 24 }}>
        <div style={{ display: 'flex', gap: 10 }}>
          <button style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 8, padding: '16px 10px', borderRadius: 18, border: '1.5px solid var(--border)', background: 'var(--surface)', cursor: 'pointer' }}>
            <div style={{ width: 40, height: 40, borderRadius: 999, background: 'var(--pine-pale)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <MessageCircle size={18} color="var(--pine)" />
            </div>
            <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--ink)' }}>แชทสด</span>
            <span style={{ fontSize: 10, color: 'var(--ink-soft)' }}>ตลอด 24 ชม.</span>
          </button>
          <button onClick={onOpenTicket} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 8, padding: '16px 10px', borderRadius: 18, border: '1.5px solid var(--border)', background: 'var(--surface)', cursor: 'pointer' }}>
            <div style={{ width: 40, height: 40, borderRadius: 999, background: 'var(--amber-soft)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <AlertCircle size={18} color="var(--amber-ink)" />
            </div>
            <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--ink)' }}>แจ้งปัญหา</span>
            <span style={{ fontSize: 10, color: 'var(--ink-soft)' }}>สร้างตั๋วสนับสนุน</span>
          </button>
        </div>

        {tickets.length > 0 && (
          <div>
            <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 10 }}>ตั๋วสนับสนุนของฉัน</p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {tickets.map((t) => (
                <div key={t.id} style={{ display: 'flex', alignItems: 'center', gap: 12, border: '1px solid var(--border)', borderRadius: 14, padding: 12 }}>
                  <div style={{ width: 34, height: 34, borderRadius: 999, background: 'var(--pine-pale)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    <FileText size={15} color="var(--pine)" />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontSize: 12.5, fontWeight: 700, color: 'var(--ink)' }}>{t.subject}</p>
                    <p style={{ fontSize: 10.5, color: 'var(--ink-soft)' }}>{t.date}</p>
                  </div>
                  <span style={{ fontSize: 10.5, fontWeight: 700, padding: '4px 10px', borderRadius: 999, background: 'var(--amber-soft)', color: 'var(--amber-ink)', flexShrink: 0 }}>รอตอบกลับ</span>
                </div>
              ))}
            </div>
          </div>
        )}

        <div>
          <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 10 }}>คำถามที่พบบ่อย</p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {FAQS.map((f) => (
              <FaqItem key={f.id} faq={f} open={openFaqId === f.id} onToggle={() => setOpenFaqId(openFaqId === f.id ? null : f.id)} />
            ))}
          </div>
        </div>

        <div style={{ borderRadius: 16, background: 'var(--pine-pale)', padding: 14, display: 'flex', alignItems: 'center', gap: 10 }}>
          <Phone size={16} color="var(--pine)" />
          <span style={{ fontSize: 12, color: 'var(--pine)', fontWeight: 600 }}>สายด่วนฉุกเฉิน: 02-123-4567 (24 ชม.)</span>
        </div>
      </div>
    </div>
  );
}

function SupportTicketScreen({ subject, setSubject, message, setMessage, onBack, onSubmit }) {
  const canSubmit = subject.trim().length > 0 && message.trim().length > 0;
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="แจ้งปัญหา" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 18 }}>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>หัวข้อปัญหา</p>
          <TextField placeholder="เช่น ปัญหาการชำระเงิน, การจองผิดพลาด" value={subject} onChange={(e) => setSubject(e.target.value)} />
        </div>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 8 }}>รายละเอียด</p>
          <TextArea placeholder="อธิบายปัญหาที่พบโดยละเอียด เพื่อให้ทีมงานช่วยเหลือได้รวดเร็วขึ้น..." value={message} onChange={(e) => setMessage(e.target.value)} />
        </div>
        <p style={{ fontSize: 11.5, color: 'var(--ink-soft)', lineHeight: 1.6 }}>ทีมงานจะตอบกลับภายใน 24 ชั่วโมง ผ่านการแจ้งเตือนในแอปและอีเมลที่ลงทะเบียนไว้</p>
      </div>
      <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        <PrimaryButton onClick={onSubmit} disabled={!canSubmit}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><Send size={15} /> ส่งตั๋วสนับสนุน</span>
        </PrimaryButton>
      </div>
    </div>
  );
}

function Toggle({ checked, onChange, label }) {
  return (
    <button role="switch" aria-checked={checked} aria-label={label} onClick={() => onChange(!checked)}
      style={{ width: 46, height: 28, borderRadius: 999, border: 'none', padding: 3, cursor: 'pointer', flexShrink: 0, background: checked ? 'var(--pine)' : 'var(--disabled)', display: 'flex', justifyContent: checked ? 'flex-end' : 'flex-start', transition: 'background .2s' }}>
      <span style={{ width: 22, height: 22, borderRadius: 999, background: '#fff', display: 'block' }} />
    </button>
  );
}

function Toast({ message }) {
  if (!message) return null;
  return (
    <div role="status" className="fade-up" style={{ position: 'absolute', left: 20, right: 20, bottom: 92, zIndex: 50, background: 'var(--ink)', color: 'var(--paper)', padding: '12px 16px', borderRadius: 14, fontSize: 12.5, fontWeight: 600, textAlign: 'center', boxShadow: '0 10px 24px -8px rgba(0,0,0,0.4)' }}>
      {message}
    </div>
  );
}

const BOOKING_STATUS = {
  confirmed: { label: 'ยืนยันแล้ว', bg: 'var(--pine-pale)', color: 'var(--pine)' },
  completed: { label: 'เสร็จสิ้น', bg: 'var(--muted-bg)', color: 'var(--ink-soft)' },
  cancelled: { label: 'ยกเลิกแล้ว', bg: 'var(--danger-soft)', color: 'var(--danger)' },
};
const CANCEL_REASONS = ['เปลี่ยนแผนการเดินทาง', 'จองผิดวันหรือผิดรายการ', 'พบราคาที่ถูกกว่า', 'เหตุฉุกเฉิน', 'อื่นๆ'];

function refundFor(booking) {
  const rate = booking.type === 'flight' ? 0.8 : 1;
  return Math.round(booking.amount * rate);
}

function StatusPill({ status }) {
  const m = BOOKING_STATUS[status];
  return <span style={{ fontSize: 10.5, fontWeight: 700, padding: '4px 10px', borderRadius: 999, background: m.bg, color: m.color, flexShrink: 0 }}>{m.label}</span>;
}

function BookingsScreen({ bookings, onBack, onSelect, onStartSearch }) {
  const [tab, setTab] = useState('all');
  const tabs = [
    { id: 'all', label: 'ทั้งหมด' },
    { id: 'confirmed', label: 'ที่จะถึง' },
    { id: 'completed', label: 'เสร็จสิ้น' },
    { id: 'cancelled', label: 'ยกเลิก' },
  ];
  const list = bookings.filter((b) => tab === 'all' || b.status === tab);
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="การจองของฉัน" onBack={onBack} />
      <div className="scrollbar-hide" style={{ padding: '14px 20px 0', display: 'flex', gap: 8, overflowX: 'auto', flexShrink: 0 }}>
        {tabs.map((t) => <Chip key={t.id} label={t.label} active={tab === t.id} onClick={() => setTab(t.id)} />)}
      </div>
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: '16px 20px 20px', display: 'flex', flexDirection: 'column', gap: 12 }}>
        {list.length === 0 && (
          <div style={{ textAlign: 'center', marginTop: 48, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 12 }}>
            <p style={{ fontSize: 13, color: 'var(--ink-soft)' }}>ยังไม่มีการจองในหมวดนี้</p>
            {tab === 'all' && <div style={{ width: 160 }}><PrimaryButton onClick={onStartSearch}>ค้นหาและจอง</PrimaryButton></div>}
          </div>
        )}
        {list.map((b) => {
          const Icon = TYPE_ICON[b.type] || Sparkles;
          return (
            <button key={b.id} onClick={() => onSelect(b.id)} style={{ display: 'flex', gap: 14, textAlign: 'left', border: '1px solid var(--border)', borderRadius: 18, padding: 14, background: 'var(--surface)', cursor: 'pointer' }}>
              <div style={{ width: 48, height: 48, borderRadius: 14, background: TYPE_GRADIENT[b.type] || TYPE_GRADIENT.trip, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <Icon size={20} color="#fff" />
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8, alignItems: 'flex-start' }}>
                  <p style={{ fontSize: 13.5, fontWeight: 700, color: 'var(--ink)' }}>{b.title}</p>
                  <StatusPill status={b.status} />
                </div>
                <p style={{ fontSize: 11.5, color: 'var(--ink-soft)', margin: '3px 0 8px' }}>{b.subtitle}</p>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: 11, color: 'var(--ink-soft)' }}>{b.number} · {b.date}</span>
                  <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--pine)' }}>฿{b.amount.toLocaleString()}</span>
                </div>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

function BookingDetailScreen({ booking, onBack, onCancel, onReview }) {
  if (!booking) return null;
  const Icon = TYPE_ICON[booking.type] || Sparkles;
  const cancelled = booking.status === 'cancelled';
  const steps = [
    { label: 'สร้างการจอง', done: true },
    { label: 'ชำระเงินแล้ว', done: true },
    { label: 'ยืนยันการจอง', done: true },
    cancelled
      ? { label: 'ยกเลิกการจอง', done: true, danger: true }
      : { label: 'เสร็จสิ้น', done: booking.status === 'completed' },
  ];
  const rows = [
    ['หมายเลขการจอง', booking.number],
    ['วันที่จอง', booking.date],
    ['วิธีชำระเงิน', booking.method],
    ['ยอดชำระ', `฿${booking.amount.toLocaleString()}`],
  ];
  if (cancelled) {
    rows.push(['เหตุผลที่ยกเลิก', booking.cancelReason || '-'], ['เงินคืน', `฿${(booking.refund || 0).toLocaleString()} (เข้า Sabai Wallet)`]);
  }
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="รายละเอียดการจอง" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'flex', gap: 14, alignItems: 'center' }}>
          <div style={{ width: 56, height: 56, borderRadius: 16, background: TYPE_GRADIENT[booking.type] || TYPE_GRADIENT.trip, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
            <Icon size={24} color="#fff" />
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <p className="font-display" style={{ fontSize: 17, fontWeight: 600 }}>{booking.title}</p>
            <p style={{ fontSize: 12, color: 'var(--ink-soft)', marginTop: 2 }}>{booking.subtitle}</p>
          </div>
          <StatusPill status={booking.status} />
        </div>

        <div style={{ border: '1px solid var(--border)', borderRadius: 18, overflow: 'hidden' }}>
          {rows.map(([k, v], i) => (
            <div key={k} style={{ display: 'flex', justifyContent: 'space-between', gap: 12, padding: '12px 16px', borderBottom: i < rows.length - 1 ? '1px solid var(--border)' : 'none', fontSize: 12.5 }}>
              <span style={{ color: 'var(--ink-soft)' }}>{k}</span>
              <span style={{ fontWeight: 600, color: 'var(--ink)', textAlign: 'right' }}>{v}</span>
            </div>
          ))}
        </div>

        <div>
          <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 12 }}>ความคืบหน้า</p>
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            {steps.map((s, i) => (
              <div key={i} style={{ display: 'flex', gap: 12 }}>
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                  <div style={{ width: 10, height: 10, borderRadius: 999, marginTop: 4, background: s.done ? (s.danger ? 'var(--danger)' : 'var(--pine)') : 'var(--border)' }} />
                  {i < steps.length - 1 && <div style={{ width: 1.5, flex: 1, background: 'var(--border)', minHeight: 22 }} />}
                </div>
                <p style={{ fontSize: 12.5, paddingBottom: 14, fontWeight: s.done ? 600 : 400, color: s.done ? 'var(--ink)' : 'var(--ink-soft)' }}>{s.label}</p>
              </div>
            ))}
          </div>
        </div>
      </div>
      {(booking.status === 'confirmed' || (booking.status === 'completed' && onReview)) && (
        <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
          {booking.status === 'confirmed'
            ? <SecondaryButton onClick={onCancel}>ยกเลิกการจอง</SecondaryButton>
            : <PrimaryButton onClick={onReview}>เขียนรีวิว</PrimaryButton>}
        </div>
      )}
    </div>
  );
}

function CancelBookingScreen({ booking, reason, setReason, onBack, onConfirm }) {
  if (!booking) return null;
  const refund = refundFor(booking);
  const pct = booking.type === 'flight' ? 80 : 100;
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="ยกเลิกการจอง" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ border: '1px solid var(--border)', borderRadius: 16, padding: 14 }}>
          <p style={{ fontSize: 13.5, fontWeight: 700 }}>{booking.title}</p>
          <p style={{ fontSize: 11.5, color: 'var(--ink-soft)', marginTop: 2 }}>{booking.number}</p>
        </div>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>เหตุผลที่ยกเลิก</p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {CANCEL_REASONS.map((r) => <Chip key={r} label={r} active={reason === r} onClick={() => setReason(r)} />)}
          </div>
        </div>
        <div style={{ borderRadius: 16, background: 'var(--pine-pale)', padding: 16 }}>
          <p style={{ fontSize: 12, color: 'var(--ink-soft)' }}>เงินคืนที่คุณจะได้รับ ({pct}%)</p>
          <p className="font-display" style={{ fontSize: 24, fontWeight: 700, color: 'var(--pine)', margin: '4px 0' }}>฿{refund.toLocaleString()}</p>
          <p style={{ fontSize: 11.5, color: 'var(--ink-soft)', lineHeight: 1.5 }}>เงินคืนเข้า Sabai Wallet ทันทีหลังยืนยันการยกเลิก{booking.type === 'flight' ? ' (ตั๋วเครื่องบินหักค่าธรรมเนียมยกเลิก 20%)' : ''}</p>
        </div>
      </div>
      <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        <PrimaryButton onClick={onConfirm} disabled={!reason}>ยืนยันการยกเลิก</PrimaryButton>
      </div>
    </div>
  );
}

function SettingRow({ icon: Icon, label, hint, right, onClick, last }) {
  const inner = (
    <>
      <Icon size={17} color="var(--pine)" style={{ flexShrink: 0 }} />
      <div style={{ flex: 1, minWidth: 0, textAlign: 'left' }}>
        <p style={{ fontSize: 13.5, color: 'var(--ink)' }}>{label}</p>
        {hint && <p style={{ fontSize: 11, color: 'var(--ink-soft)', marginTop: 1 }}>{hint}</p>}
      </div>
      {right}
    </>
  );
  const style = { display: 'flex', alignItems: 'center', gap: 12, padding: '13px 16px', borderBottom: last ? 'none' : '1px solid var(--border)', width: '100%', background: 'none', borderTop: 'none', borderLeft: 'none', borderRight: 'none' };
  return onClick
    ? <button onClick={onClick} style={{ ...style, cursor: 'pointer' }}>{inner}</button>
    : <div style={style}>{inner}</div>;
}

function SettingsScreen({ settings, onChange, onChangePassword, onDeleteAccount, onBack }) {
  const [confirmDelete, setConfirmDelete] = useState(false);
  const group = (title, children) => (
    <div>
      <p style={{ fontSize: 12, fontWeight: 700, color: 'var(--ink-soft)', marginBottom: 8, paddingLeft: 4 }}>{title}</p>
      <div style={{ border: '1px solid var(--border)', borderRadius: 18, overflow: 'hidden', background: 'var(--surface)' }}>{children}</div>
    </div>
  );
  const chips = (key, options) => (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, padding: '0 16px 14px' }}>
      {options.map((o) => <Chip key={o} label={o} active={settings[key] === o} onClick={() => onChange({ [key]: o })} />)}
    </div>
  );
  const notif = (key, label, hint, last) => (
    <SettingRow key={key} icon={Bell} label={label} hint={hint} last={last}
      right={<Toggle checked={settings.notif[key]} label={label} onChange={(v) => onChange({ notif: { ...settings.notif, [key]: v } })} />} />
  );
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden', position: 'relative' }}>
      <TopBar title="การตั้งค่า" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 22 }}>
        {group('ทั่วไป', (
          <>
            <SettingRow icon={Globe} label="ภาษา" last />
            {chips('language', ['ไทย', 'English', '中文', '日本語'])}
            <SettingRow icon={Banknote} label="สกุลเงิน" last />
            {chips('currency', ['THB ฿', 'USD $', 'JPY ¥'])}
            <SettingRow icon={Palette} label="ธีม" last />
            {chips('theme', ['สว่าง', 'มืด', 'ตามระบบ'])}
          </>
        ))}
        {group('การแจ้งเตือน', (
          <>
            {notif('push', 'การแจ้งเตือนในแอป', 'การจอง การชำระเงิน และข่าวสารบัญชี')}
            {notif('email', 'อีเมล', 'ใบยืนยันการจองและใบเสร็จ')}
            {notif('trip', 'เตือนก่อนเดินทาง', 'แจ้งเตือนล่วงหน้า 3 วัน')}
            {notif('promo', 'โปรโมชั่นและคูปอง', 'ข้อเสนอพิเศษที่เหมาะกับคุณ', true)}
          </>
        ))}
        {group('ความปลอดภัย', (
          <SettingRow icon={KeyRound} label="เปลี่ยนรหัสผ่าน" onClick={onChangePassword} last right={<ChevronRight size={16} color="var(--ink-soft)" />} />
        ))}
        <button onClick={() => setConfirmDelete(true)} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8, padding: 14, borderRadius: 16, border: '1.5px solid var(--danger)', background: 'transparent', color: 'var(--danger)', fontWeight: 700, fontSize: 13.5, cursor: 'pointer' }}>
          <Trash2 size={16} /> ลบบัญชี
        </button>
      </div>
      {confirmDelete && (
        <div role="dialog" aria-modal="true" aria-label="ยืนยันการลบบัญชี" style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 28, zIndex: 40 }}>
          <div style={{ background: 'var(--surface)', borderRadius: 20, padding: 22, width: '100%', display: 'flex', flexDirection: 'column', gap: 14 }}>
            <p className="font-display" style={{ fontSize: 17, fontWeight: 600 }}>ลบบัญชีถาวร?</p>
            <p style={{ fontSize: 12.5, color: 'var(--ink-soft)', lineHeight: 1.6 }}>ข้อมูลทริป การจอง คูปอง และยอด Sabai Wallet จะถูกลบและกู้คืนไม่ได้</p>
            <div style={{ display: 'flex', gap: 10 }}>
              <div style={{ flex: 1 }}><SecondaryButton onClick={() => setConfirmDelete(false)}>ยกเลิก</SecondaryButton></div>
              <button onClick={onDeleteAccount} style={{ flex: 1, padding: 15, borderRadius: 16, border: 'none', background: 'var(--danger)', color: '#fff', fontWeight: 700, fontSize: 15, cursor: 'pointer' }}>ลบบัญชี</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function ChangePasswordScreen({ onBack, onSubmit }) {
  const [current, setCurrent] = useState('');
  const [next, setNext] = useState('');
  const [confirm, setConfirm] = useState('');
  const errors = [];
  if (next && next.length < 8) errors.push('รหัสผ่านใหม่ต้องมีอย่างน้อย 8 ตัวอักษร');
  if (next && current && next === current) errors.push('รหัสผ่านใหม่ต้องไม่ซ้ำกับรหัสผ่านเดิม');
  if (confirm && next !== confirm) errors.push('ยืนยันรหัสผ่านไม่ตรงกัน');
  const valid = current && next.length >= 8 && next === confirm && next !== current;
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="เปลี่ยนรหัสผ่าน" onBack={onBack} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 14 }}>
        <TextField icon={Lock} type="password" placeholder="รหัสผ่านปัจจุบัน" value={current} onChange={(e) => setCurrent(e.target.value)} />
        <TextField icon={Lock} type="password" placeholder="รหัสผ่านใหม่ (อย่างน้อย 8 ตัว)" value={next} onChange={(e) => setNext(e.target.value)} />
        <TextField icon={Lock} type="password" placeholder="ยืนยันรหัสผ่านใหม่" value={confirm} onChange={(e) => setConfirm(e.target.value)} />
        {errors.map((e) => <p key={e} style={{ fontSize: 11.5, color: 'var(--danger)' }}>{e}</p>)}
      </div>
      <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        <PrimaryButton onClick={onSubmit} disabled={!valid}>บันทึกรหัสผ่านใหม่</PrimaryButton>
      </div>
    </div>
  );
}

/* ---------------------------------------------------------------- */
/*  COMMUNITY                                                         */
/* ---------------------------------------------------------------- */

const ME_ID = 'me';
const POST_MAX = 500;
const POST_MIN = 10;

const MEMBERS = [
  { id: 'me', name: 'หนาน ใจดี', initial: 'น', color: '#10403D', level: 'นักเดินทางใหม่', bio: 'ชอบเที่ยวธรรมชาติและตามหาคาเฟ่เงียบๆ', followers: 12 },
  { id: 'm1', name: 'มะลิ ทองดี', initial: 'ม', color: '#B5651D', level: 'Local Guide เชียงใหม่', bio: 'อยู่เชียงใหม่มา 8 ปี พาเที่ยวมุมที่นักท่องเที่ยวไม่ค่อยรู้จัก', followers: 2840 },
  { id: 'm2', name: 'ต้น สายลุย', initial: 'ต', color: '#123C69', level: 'นักเดินทางสายลุย', bio: 'เดินป่า ดำน้ำ ปั่นจักรยาน งบน้อยแต่ไปได้ไกล', followers: 1530 },
  { id: 'm3', name: 'ฝ้าย คาเฟ่ฮอป', initial: 'ฝ', color: '#7A3B6B', level: 'นักชิมตัวยง', bio: 'รีวิวคาเฟ่และร้านอาหารท้องถิ่นทั่วไทย', followers: 4120 },
  { id: 'm4', name: 'ภูมิ ทริปครอบครัว', initial: 'ภ', color: '#0E6E55', level: 'Local Guide ภูเก็ต', bio: 'พ่อลูกสองที่ชอบวางแผนทริปให้ทุกคนสนุก', followers: 960 },
];

const GROUP_RULES = [
  'พูดจาสุภาพและให้เกียรติสมาชิกทุกคน',
  'ห้ามโฆษณาหรือขายสินค้าโดยไม่ได้รับอนุญาตจากผู้ดูแลกลุ่ม',
  'แชร์ข้อมูลจากประสบการณ์จริง และระบุเดือนที่ไปเมื่อพูดถึงราคา',
];

const GROUPS = [
  { id: 'g1', name: 'เที่ยวเชียงใหม่ไม่ซ้ำใคร', desc: 'แลกเปลี่ยนที่เที่ยว ร้านอาหาร และเส้นทางในเชียงใหม่และจังหวัดใกล้เคียง', destination: 'เชียงใหม่', members: 12840, gradient: 'linear-gradient(135deg,#1C5C57,#89A66B)', rules: GROUP_RULES },
  { id: 'g2', name: 'สายคาเฟ่และของกิน', desc: 'รวมคาเฟ่และร้านอาหารท้องถิ่นที่อยากแนะนำ พร้อมรูปและเมนูเด็ด', destination: null, members: 9320, gradient: 'linear-gradient(135deg,#B5651D,#7A4B22)', rules: GROUP_RULES },
  { id: 'g3', name: 'เที่ยวทะเลงบประหยัด', desc: 'เคล็ดลับเที่ยวทะเลอันดามันและอ่าวไทยแบบไม่เสียเงินเกินจำเป็น', destination: 'ภูเก็ต', members: 7410, gradient: 'linear-gradient(135deg,#0B4F6C,#01BAEF)', rules: GROUP_RULES },
  { id: 'g4', name: 'เที่ยวคนเดียว', desc: 'พื้นที่สำหรับคนที่ชอบเดินทางคนเดียว แชร์ความปลอดภัย ที่พัก และการเดินทาง', destination: null, members: 5280, gradient: 'linear-gradient(135deg,#3B2352,#A768C9)', rules: GROUP_RULES },
];

const POST_TYPE_META = {
  story: { label: 'เรื่องเล่า', icon: BookOpen },
  tip: { label: 'เคล็ดลับ', icon: Lightbulb },
  question: { label: 'คำถาม', icon: HelpCircle },
  trip: { label: 'ทริปแนะนำ', icon: Briefcase },
};
const COMPOSER_TYPES = ['story', 'tip', 'question'];
const COMPOSER_HINT = {
  story: 'เล่าประสบการณ์การเดินทางของคุณ เช่น สถานที่ที่ไป สิ่งที่ประทับใจ',
  tip: 'แบ่งปันเคล็ดลับที่ช่วยให้คนอื่นเที่ยวง่ายขึ้น เช่น เวลาที่ควรไป วิธีเดินทาง',
  question: 'ถามสิ่งที่อยากรู้ เช่น ที่พักแถวไหนดี ต้องจองล่วงหน้าไหม',
};
const REPORT_REASONS = ['สแปมหรือโฆษณา', 'ข้อมูลเท็จหรือทำให้เข้าใจผิด', 'ภาษาไม่สุภาพหรือคุกคาม', 'ละเมิดความเป็นส่วนตัว', 'เหตุผลอื่น'];

const INITIAL_POSTS = [
  { id: 'p1', authorId: 'm1', type: 'tip', destination: 'เชียงใหม่', groupId: 'g1', time: '2 ชม.', likes: 248, text: 'ถ้าจะขึ้นดอยสุเทพ แนะนำให้ไปถึงก่อน 7 โมงเช้า รถยังไม่ติด อากาศเย็น และมีโอกาสเจอทะเลหมกด้านล่าง ลงมาแล้วแวะกินข้าวซอยร้านแถวเมืองเก่าต่อได้พอดี' },
  { id: 'p2', authorId: 'm3', type: 'story', destination: 'เชียงใหม่', groupId: 'g2', time: '5 ชม.', likes: 512, hasImage: true, text: 'เจอคาเฟ่ริมนาที่แม่ริม นั่งมองทุ่งข้าวได้ทั้งบ่าย เครื่องดื่มรสละมุน ขนมโฮมเมดอร่อย บรรยากาศเงียบ เหมาะกับคนที่อยากพักจากความวุ่นวายในเมือง' },
  { id: 'p3', authorId: 'm2', type: 'trip', destination: 'ภูเก็ต', groupId: 'g3', time: 'เมื่อวาน', likes: 186, sharedTrip: { days: 4, budget: 12000, styles: ['ธรรมชาติ', 'ผจญภัย'] }, text: 'สรุปทริปภูเก็ต 4 วันแบบไม่ใช้ทัวร์ เช่ามอเตอร์ไซค์เที่ยวหาดทางใต้ ดูพระอาทิตย์ตกที่แหลมพรหมเทพ และกินอาหารทะเลตลาดท้องถิ่น' },
  { id: 'p4', authorId: 'm4', type: 'question', destination: 'ภูเก็ต', groupId: 'g3', time: 'เมื่อวาน', likes: 34, text: 'จะพาลูกอายุ 5 ขวบไปภูเก็ตช่วงพฤศจิกายน ควรพักหาดไหนดีคะ อยากได้หาดที่คลื่นไม่แรงและมีร้านอาหารใกล้ๆ' },
  { id: 'p5', authorId: 'm2', type: 'question', destination: 'กระบี่', groupId: null, time: '3 ชม.', likes: 9, text: 'ไปเกาะพีพีจากอ่าวนางต้องจองทัวร์ล่วงหน้าไหม หรือไปซื้อหน้างานได้เลย' },
  { id: 'p6', authorId: 'm1', type: 'trip', destination: 'เชียงใหม่', groupId: 'g1', time: '2 วัน', likes: 421, sharedTrip: { days: 3, budget: 10000, styles: ['ธรรมชาติ', 'อาหารท้องถิ่น'] }, text: 'แผนเที่ยวเชียงใหม่ 3 วัน เมืองเก่า ดอยอินทนนท์ และคาเฟ่ย่านนิมมาน เหมาะกับคู่รักที่ชอบธรรมชาติและของกิน' },
  { id: 'p7', authorId: 'm3', type: 'tip', destination: 'กรุงเทพฯ', groupId: 'g2', time: '3 วัน', likes: 97, text: 'เดินเที่ยวย่านเจริญกรุงช่วงเย็นกำลังดี แดดอ่อน ร้านอาหารและแกลเลอรีเปิดหลายที่ ใช้เรือข้ามฟากสลับกับเดินจะประหยัดเวลากว่านั่งรถ' },
  { id: 'p8', authorId: 'me', type: 'story', destination: 'ปาย', groupId: null, time: '1 สัปดาห์', likes: 23, text: 'ไปปายครั้งแรกแล้วติดใจ ตื่นเช้ามานั่งจิบกาแฟมองหมอกบนเขา เดินเล่นถนนคนเดินตอนเย็น ช้าๆ แต่ชีวิตดีมาก' },
];

const INITIAL_COMMENTS = {
  p1: [
    { id: 'c1', authorId: 'm2', text: 'ขอบคุณค่ะ ไปมาแล้วจริง ไปสายหน่อยรถติดมาก', time: '1 ชม.', likes: 6 },
    { id: 'c2', authorId: 'm3', text: 'ลงมาแล้วแนะนำร้านข้าวซอยที่อยู่ในซอยเล็กๆ ใกล้ประตูช้างเผือก', time: '45 นาที', likes: 3 },
  ],
  p2: [{ id: 'c3', authorId: 'm1', text: 'ชอบมากค่ะ ช่วงเย็นแสงสวยสุด', time: '3 ชม.', likes: 4 }],
  p4: [
    { id: 'c4', authorId: 'm1', text: 'ลองหาดกะตะหรือหาดในหาน ช่วงพฤศจิกายนทะเลเริ่มสงบ เด็กเล็กเล่นน้ำตื้นได้สบาย', time: '20 ชม.', likes: 11, best: true },
    { id: 'c5', authorId: 'm2', text: 'หาดกะรนก็ดี ร้านอาหารเยอะ เดินจากที่พักมาได้', time: '18 ชม.', likes: 5 },
    { id: 'c6', authorId: 'm3', text: 'แนะนำดูสภาพคลื่นล่วงหน้า บางวันมีธงแดง', time: '15 ชม.', likes: 2 },
  ],
  p6: [{ id: 'c7', authorId: 'm4', text: 'เซฟไว้เลย ขอบคุณที่แชร์แผนละเอียดนะคะ', time: '1 วัน', likes: 2 }],
};

const formatCount = (n) => (n >= 1000 ? `${(n / 1000).toFixed(1).replace(/\.0$/, '')}k` : String(n));
const destinationGradient = (name) => (DESTINATIONS.find((d) => d.name === name) || {}).gradient || 'linear-gradient(135deg,#1C5C57,#0B2422)';

function useCommunity({ showToast, pushNotification, onSaveTrip }) {
  const [posts, setPosts] = useState(INITIAL_POSTS);
  const [comments, setComments] = useState(INITIAL_COMMENTS);
  const [liked, setLiked] = useState([]);
  const [saved, setSaved] = useState([]);
  const [likedComments, setLikedComments] = useState([]);
  const [following, setFollowing] = useState(['m1']);
  const [joined, setJoined] = useState(['g1']);
  const [hidden, setHidden] = useState([]);
  const [copiedTrips, setCopiedTrips] = useState([]);
  const [ui, setUi] = useState({ tab: 'feed', feedMode: 'all', dest: 'ทั้งหมด', qaFilter: 'all' });
  const [selected, setSelected] = useState({ postId: null, memberId: null, groupId: null });
  const [composer, setComposer] = useState({ type: 'story', destination: '', text: '', groupId: null });
  const [sheet, setSheet] = useState(null);
  const timers = useRef([]);
  useEffect(() => () => timers.current.forEach(clearTimeout), []);

  const toggle = (setter, id) => setter((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));

  const memberOf = (id) => MEMBERS.find((m) => m.id === id) || MEMBERS[0];
  const commentsOf = (postId) => comments[postId] || [];
  const visiblePosts = posts.filter((p) => !hidden.includes(p.id));
  const postById = (id) => posts.find((p) => p.id === id) || null;
  const likeCount = (post) => post.likes + (liked.includes(post.id) ? 1 : 0);
  const followerCount = (m) => m.followers + (following.includes(m.id) ? 1 : 0);
  const memberCount = (g) => g.members + (joined.includes(g.id) ? 1 : 0);

  const toggleLike = (id) => toggle(setLiked, id);
  const toggleSave = (id) => {
    const wasSaved = saved.includes(id);
    toggle(setSaved, id);
    showToast(wasSaved ? 'เลิกบันทึกโพสต์แล้ว' : 'บันทึกโพสต์แล้ว');
  };
  const toggleCommentLike = (id) => toggle(setLikedComments, id);
  const toggleFollow = (id) => {
    const was = following.includes(id);
    toggle(setFollowing, id);
    showToast(was ? `เลิกติดตาม ${memberOf(id).name} แล้ว` : `ติดตาม ${memberOf(id).name} แล้ว`);
  };
  const toggleJoin = (id) => {
    const was = joined.includes(id);
    toggle(setJoined, id);
    const g = GROUPS.find((x) => x.id === id);
    showToast(was ? `ออกจากกลุ่ม "${g.name}" แล้ว` : `เข้าร่วมกลุ่ม "${g.name}" แล้ว`);
  };

  const addComment = (postId, text) => {
    const clean = text.trim();
    if (!clean) return;
    setComments((prev) => ({ ...prev, [postId]: [...(prev[postId] || []), { id: `c${Date.now()}`, authorId: ME_ID, text: clean, time: 'เมื่อสักครู่', likes: 0 }] }));
  };
  const markBest = (postId, commentId) => {
    setComments((prev) => ({ ...prev, [postId]: (prev[postId] || []).map((c) => ({ ...c, best: c.id === commentId ? !c.best : false })) }));
    showToast('อัปเดตคำตอบที่ดีที่สุดแล้ว');
  };

  const startPost = ({ type = 'story', groupId = null } = {}) => {
    const g = GROUPS.find((x) => x.id === groupId);
    setComposer({ type, destination: (g && g.destination) || '', text: '', groupId });
  };
  const updateComposer = (patch) => setComposer((prev) => ({ ...prev, ...patch }));
  const createPost = () => {
    const text = composer.text.trim();
    if (text.length < POST_MIN || text.length > POST_MAX || !composer.destination) return null;
    const id = `p${Date.now()}`;
    setPosts((prev) => [{ id, authorId: ME_ID, type: composer.type, destination: composer.destination, groupId: composer.groupId, time: 'เมื่อสักครู่', likes: 0, text }, ...prev]);
    setUi((prev) => ({ ...prev, tab: composer.type === 'question' ? 'qa' : 'feed', feedMode: 'all', dest: 'ทั้งหมด' }));
    showToast('โพสต์ของคุณเผยแพร่แล้ว');
    // จำลองปฏิกิริยาจากสมาชิกคนอื่นเพื่อแสดงการแจ้งเตือนของชุมชน
    timers.current.push(setTimeout(() => {
      setPosts((prev) => prev.map((p) => (p.id === id ? { ...p, likes: p.likes + 3 } : p)));
      pushNotification('community', 'มีคนถูกใจโพสต์ของคุณ', 'มะลิ ทองดี และอีก 2 คนถูกใจโพสต์ของคุณในชุมชน');
    }, 4000));
    setComposer({ type: 'story', destination: '', text: '', groupId: null });
    return id;
  };
  const deletePost = (id) => {
    setPosts((prev) => prev.filter((p) => p.id !== id));
    setSaved((prev) => prev.filter((x) => x !== id));
    showToast('ลบโพสต์แล้ว');
  };
  const reportPost = (id) => {
    setHidden((prev) => [...prev, id]);
    showToast('ส่งรายงานแล้ว ขอบคุณที่ช่วยดูแลชุมชน');
  };
  const hidePost = (id) => {
    setHidden((prev) => [...prev, id]);
    showToast('ซ่อนโพสต์นี้แล้ว');
  };
  const sharePost = async (id) => {
    try {
      await navigator.clipboard.writeText(`https://sabai.example/community/${id}`);
      showToast('คัดลอกลิงก์โพสต์แล้ว');
    } catch (e) {
      showToast('คัดลอกลิงก์ไม่สำเร็จ ลองอีกครั้ง');
    }
  };
  const copyTrip = (post) => {
    if (!post.sharedTrip) return;
    if (copiedTrips.includes(post.id)) { showToast('คุณบันทึกทริปนี้ไว้แล้ว'); return; }
    onSaveTrip(post);
    setCopiedTrips((prev) => [...prev, post.id]);
    showToast('บันทึกทริปแล้ว ดูได้ที่แท็บทริป');
  };

  return {
    posts, visiblePosts, comments, liked, saved, likedComments, following, joined, copiedTrips, ui, selected, composer, sheet,
    memberOf, commentsOf, postById, likeCount, followerCount, memberCount,
    setUi: (patch) => setUi((prev) => ({ ...prev, ...patch })),
    selectPost: (postId) => setSelected((prev) => ({ ...prev, postId })),
    selectMember: (memberId) => setSelected((prev) => ({ ...prev, memberId })),
    selectGroup: (groupId) => setSelected((prev) => ({ ...prev, groupId })),
    openSheet: setSheet, closeSheet: () => setSheet(null),
    toggleLike, toggleSave, toggleCommentLike, toggleFollow, toggleJoin,
    addComment, markBest, startPost, updateComposer, createPost, deletePost, reportPost, hidePost, sharePost, copyTrip,
  };
}

function Avatar({ member, size = 38 }) {
  return (
    <div aria-hidden="true" style={{ width: size, height: size, borderRadius: 999, background: member.color, color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, fontSize: Math.round(size * 0.42), flexShrink: 0 }}>
      {member.initial}
    </div>
  );
}

function PostAction({ icon: Icon, label, count, pressed, color, onClick, fill }) {
  const isToggle = typeof pressed === 'boolean';
  return (
    <button onClick={onClick} aria-label={label} aria-pressed={isToggle ? pressed : undefined}
      style={{ display: 'flex', alignItems: 'center', gap: 5, background: 'none', border: 'none', cursor: 'pointer', padding: '6px 4px', color: pressed ? color : 'var(--ink-soft)', fontSize: 12, fontWeight: 600 }}>
      <Icon size={18} fill={pressed && fill ? color : 'none'} />
      {count !== undefined && <span>{formatCount(count)}</span>}
    </button>
  );
}

function PostCard({ post, cm, nav, detail = false, linkAuthor = true }) {
  const author = cm.memberOf(post.authorId);
  const meta = POST_TYPE_META[post.type];
  const TypeIcon = meta.icon;
  const group = post.groupId ? GROUPS.find((g) => g.id === post.groupId) : null;
  const postComments = cm.commentsOf(post.id);
  const isLiked = cm.liked.includes(post.id);
  const isSaved = cm.saved.includes(post.id);
  const hasBest = postComments.some((c) => c.best);
  const tripCopied = cm.copiedTrips.includes(post.id);
  const openPost = () => { if (!detail) { cm.selectPost(post.id); nav.push('post-detail'); } };
  const openAuthor = () => { cm.selectMember(author.id); nav.push('member'); };
  const clamp = detail ? {} : { display: '-webkit-box', WebkitLineClamp: 4, WebkitBoxOrient: 'vertical', overflow: 'hidden' };
  return (
    <article style={{ border: '1px solid var(--border)', borderRadius: 18, background: 'var(--surface)', padding: 14, display: 'flex', flexDirection: 'column', gap: 10 }}>
      <header style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <button onClick={linkAuthor ? openAuthor : undefined} aria-label={`ดูโปรไฟล์ ${author.name}`} disabled={!linkAuthor}
          style={{ display: 'flex', alignItems: 'center', gap: 10, flex: 1, minWidth: 0, background: 'none', border: 'none', padding: 0, cursor: linkAuthor ? 'pointer' : 'default', textAlign: 'left' }}>
          <Avatar member={author} />
          <span style={{ minWidth: 0 }}>
            <span style={{ display: 'block', fontSize: 13, fontWeight: 700, color: 'var(--ink)' }}>{author.name}</span>
            <span style={{ display: 'block', fontSize: 11, color: 'var(--ink-soft)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              {author.level} ({post.time})
            </span>
          </span>
        </button>
        <button onClick={() => cm.openSheet({ type: 'post', postId: post.id })} aria-label="ตัวเลือกโพสต์" style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 6, display: 'flex' }}>
          <MoreHorizontal size={18} color="var(--ink-soft)" />
        </button>
      </header>

      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 11, fontWeight: 700, padding: '4px 10px', borderRadius: 999, background: 'var(--pine-pale)', color: 'var(--pine)' }}>
          <TypeIcon size={12} /> {meta.label}
        </span>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 11, fontWeight: 600, padding: '4px 10px', borderRadius: 999, background: 'var(--muted-bg)', color: 'var(--ink-soft)' }}>
          <MapPin size={11} /> {post.destination}
        </span>
        {group && (
          <span style={{ fontSize: 11, fontWeight: 600, padding: '4px 10px', borderRadius: 999, background: 'var(--muted-bg)', color: 'var(--ink-soft)' }}>{group.name}</span>
        )}
      </div>

      {detail ? (
        <p style={{ fontSize: 14, lineHeight: 1.7, color: 'var(--ink)', whiteSpace: 'pre-wrap' }}>{post.text}</p>
      ) : (
        <button onClick={openPost} style={{ background: 'none', border: 'none', padding: 0, cursor: 'pointer', textAlign: 'left' }}>
          <span style={{ fontSize: 13.5, lineHeight: 1.65, color: 'var(--ink)', ...clamp }}>{post.text}</span>
        </button>
      )}

      {post.hasImage && (
        <div role="img" aria-label={`ภาพบรรยากาศ ${post.destination}`} style={{ height: 150, borderRadius: 14, background: destinationGradient(post.destination), display: 'flex', alignItems: 'flex-end', padding: 12 }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, color: '#fff', fontSize: 12, fontWeight: 700 }}><MapPin size={13} /> {post.destination}</span>
        </div>
      )}

      {post.sharedTrip && (
        <div style={{ border: '1px dashed var(--border)', borderRadius: 14, padding: 12, display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <p style={{ fontSize: 13, fontWeight: 700, color: 'var(--ink)' }}>ทริป{post.destination} {post.sharedTrip.days} วัน</p>
            <p style={{ fontSize: 11.5, color: 'var(--ink-soft)', marginTop: 2 }}>งบประมาณ ฿{post.sharedTrip.budget.toLocaleString()} สไตล์ {post.sharedTrip.styles.join(', ')}</p>
          </div>
          <button onClick={() => cm.copyTrip(post)} disabled={tripCopied}
            style={{ padding: '8px 12px', borderRadius: 12, border: '1.5px solid var(--pine)', background: tripCopied ? 'var(--pine-pale)' : 'transparent', color: 'var(--pine)', fontWeight: 700, fontSize: 12, cursor: tripCopied ? 'default' : 'pointer', flexShrink: 0 }}>
            {tripCopied ? 'บันทึกแล้ว' : 'บันทึกเป็นทริปของฉัน'}
          </button>
        </div>
      )}

      {post.type === 'question' && (
        <span style={{ alignSelf: 'flex-start', display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 11, fontWeight: 700, padding: '4px 10px', borderRadius: 999, background: hasBest ? 'var(--pine-pale)' : 'var(--amber-soft)', color: hasBest ? 'var(--pine)' : 'var(--amber-ink)' }}>
          {hasBest ? <CheckCircle2 size={12} /> : <Clock size={12} />}
          {hasBest ? 'มีคำตอบที่ดีที่สุด' : postComments.length > 0 ? `มี ${postComments.length} คำตอบ` : 'รอคำตอบ'}
        </span>
      )}

      <footer style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderTop: '1px solid var(--border)', paddingTop: 6 }}>
        <div style={{ display: 'flex', gap: 10 }}>
          <PostAction icon={Heart} label={isLiked ? 'ยกเลิกถูกใจ' : 'ถูกใจ'} count={cm.likeCount(post)} pressed={isLiked} color="var(--danger)" fill onClick={() => cm.toggleLike(post.id)} />
          <PostAction icon={MessageCircle} label="ดูความคิดเห็น" count={postComments.length} onClick={openPost} />
        </div>
        <div style={{ display: 'flex', gap: 6 }}>
          <PostAction icon={Bookmark} label={isSaved ? 'เลิกบันทึกโพสต์' : 'บันทึกโพสต์'} pressed={isSaved} color="var(--pine)" fill onClick={() => cm.toggleSave(post.id)} />
          <PostAction icon={Share2} label="คัดลอกลิงก์โพสต์" onClick={() => cm.sharePost(post.id)} />
        </div>
      </footer>
    </article>
  );
}

function EmptyState({ icon: Icon, title, hint, action }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 12, textAlign: 'center', padding: '40px 20px' }}>
      <div style={{ width: 60, height: 60, borderRadius: 999, background: 'var(--pine-pale)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Icon size={24} color="var(--pine)" />
      </div>
      <p style={{ fontSize: 14, fontWeight: 700, color: 'var(--ink)' }}>{title}</p>
      {hint && <p style={{ fontSize: 12, color: 'var(--ink-soft)', maxWidth: 240, lineHeight: 1.6 }}>{hint}</p>}
      {action}
    </div>
  );
}

function FollowButton({ cm, memberId, compact = false }) {
  const isFollowing = cm.following.includes(memberId);
  return (
    <button onClick={() => cm.toggleFollow(memberId)} aria-pressed={isFollowing}
      style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 5, padding: compact ? '7px 12px' : '10px 18px', borderRadius: 999, fontSize: 12, fontWeight: 700, cursor: 'pointer', border: '1.5px solid var(--pine)', background: isFollowing ? 'transparent' : 'var(--pine)', color: isFollowing ? 'var(--pine)' : '#fff' }}>
      {isFollowing ? <UserCheck size={14} /> : <UserPlus size={14} />}
      {isFollowing ? 'กำลังติดตาม' : 'ติดตาม'}
    </button>
  );
}

function SuggestedMembers({ cm, nav }) {
  const list = MEMBERS.filter((m) => m.id !== ME_ID && !cm.following.includes(m.id));
  if (list.length === 0) return null;
  return (
    <section aria-label="แนะนำให้ติดตาม">
      <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 10 }}>แนะนำให้ติดตาม</p>
      <div className="scrollbar-hide" style={{ display: 'flex', gap: 10, overflowX: 'auto' }}>
        {list.map((m) => (
          <div key={m.id} style={{ flex: '0 0 148px', border: '1px solid var(--border)', borderRadius: 16, background: 'var(--surface)', padding: 12, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6, textAlign: 'center' }}>
            <button onClick={() => { cm.selectMember(m.id); nav.push('member'); }} aria-label={`ดูโปรไฟล์ ${m.name}`} style={{ background: 'none', border: 'none', cursor: 'pointer', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6 }}>
              <Avatar member={m} size={46} />
              <span style={{ fontSize: 12.5, fontWeight: 700, color: 'var(--ink)' }}>{m.name}</span>
              <span style={{ fontSize: 10.5, color: 'var(--ink-soft)' }}>{m.level}</span>
            </button>
            <FollowButton cm={cm} memberId={m.id} compact />
          </div>
        ))}
      </div>
    </section>
  );
}

function FeedView({ cm, nav }) {
  const { feedMode, dest } = cm.ui;
  const destinations = ['ทั้งหมด', ...DESTINATIONS.map((d) => d.name)];
  let list = cm.visiblePosts;
  if (feedMode === 'following') list = list.filter((p) => cm.following.includes(p.authorId));
  if (dest !== 'ทั้งหมด') list = list.filter((p) => p.destination === dest);
  return (
    <>
      <div style={{ display: 'flex', gap: 8 }}>
        <Chip label="สำหรับคุณ" active={feedMode === 'all'} onClick={() => cm.setUi({ feedMode: 'all' })} />
        <Chip label="กำลังติดตาม" active={feedMode === 'following'} onClick={() => cm.setUi({ feedMode: 'following' })} />
      </div>
      <div className="scrollbar-hide" style={{ display: 'flex', gap: 8, overflowX: 'auto' }}>
        {destinations.map((d) => <Chip key={d} label={d} active={dest === d} onClick={() => cm.setUi({ dest: d })} />)}
      </div>
      {feedMode === 'all' && dest === 'ทั้งหมด' && <SuggestedMembers cm={cm} nav={nav} />}
      {list.length === 0 ? (
        feedMode === 'following' ? (
          <EmptyState icon={Users} title="ยังไม่มีโพสต์จากคนที่คุณติดตาม" hint="ติดตามนักเดินทางที่ชอบสไตล์เดียวกัน แล้วโพสต์ของพวกเขาจะมาอยู่ที่นี่"
            action={<div style={{ width: 200 }}><PrimaryButton onClick={() => cm.setUi({ feedMode: 'all' })}>ดูโพสต์ทั้งหมด</PrimaryButton></div>} />
        ) : (
          <EmptyState icon={BookOpen} title={`ยังไม่มีโพสต์เกี่ยวกับ${dest}`} hint="เป็นคนแรกที่แบ่งปันประสบการณ์ของคุณ"
            action={<div style={{ width: 200 }}><PrimaryButton onClick={() => { cm.startPost({}); nav.push('create-post'); }}>เขียนโพสต์</PrimaryButton></div>} />
        )
      ) : (
        list.map((p) => <PostCard key={p.id} post={p} cm={cm} nav={nav} />)
      )}
    </>
  );
}

function QaView({ cm, nav }) {
  const { qaFilter } = cm.ui;
  const questions = cm.visiblePosts.filter((p) => p.type === 'question');
  const answered = (p) => cm.commentsOf(p.id).length > 0;
  const list = questions.filter((p) => qaFilter === 'all' || (qaFilter === 'answered' ? answered(p) : !answered(p)));
  return (
    <>
      <PrimaryButton onClick={() => { cm.startPost({ type: 'question' }); nav.push('create-post'); }}>ถามคำถามใหม่</PrimaryButton>
      <div style={{ display: 'flex', gap: 8 }}>
        <Chip label="ทั้งหมด" active={qaFilter === 'all'} onClick={() => cm.setUi({ qaFilter: 'all' })} />
        <Chip label="รอคำตอบ" active={qaFilter === 'open'} onClick={() => cm.setUi({ qaFilter: 'open' })} />
        <Chip label="มีคำตอบแล้ว" active={qaFilter === 'answered'} onClick={() => cm.setUi({ qaFilter: 'answered' })} />
      </div>
      {list.length === 0
        ? <EmptyState icon={HelpCircle} title="ไม่มีคำถามในหมวดนี้" hint="ลองเปลี่ยนตัวกรอง หรือถามคำถามที่คุณสงสัย" />
        : list.map((p) => <PostCard key={p.id} post={p} cm={cm} nav={nav} />)}
    </>
  );
}

function GroupRow({ group, cm, nav }) {
  const isJoined = cm.joined.includes(group.id);
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12, border: '1px solid var(--border)', borderRadius: 16, background: 'var(--surface)', padding: 12 }}>
      <button onClick={() => { cm.selectGroup(group.id); nav.push('group-detail'); }} aria-label={`เปิดกลุ่ม ${group.name}`}
        style={{ display: 'flex', alignItems: 'center', gap: 12, flex: 1, minWidth: 0, background: 'none', border: 'none', padding: 0, cursor: 'pointer', textAlign: 'left' }}>
        <div aria-hidden="true" style={{ width: 46, height: 46, borderRadius: 14, background: group.gradient, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
          <Users size={20} color="#fff" />
        </div>
        <span style={{ minWidth: 0 }}>
          <span style={{ display: 'block', fontSize: 13.5, fontWeight: 700, color: 'var(--ink)' }}>{group.name}</span>
          <span style={{ display: 'block', fontSize: 11, color: 'var(--ink-soft)', marginTop: 2 }}>สมาชิก {cm.memberCount(group).toLocaleString()} คน</span>
        </span>
      </button>
      <button onClick={() => cm.toggleJoin(group.id)} aria-pressed={isJoined}
        style={{ padding: '8px 14px', borderRadius: 999, fontSize: 12, fontWeight: 700, cursor: 'pointer', flexShrink: 0, border: '1.5px solid var(--pine)', background: isJoined ? 'transparent' : 'var(--pine)', color: isJoined ? 'var(--pine)' : '#fff' }}>
        {isJoined ? 'เข้าร่วมแล้ว' : 'เข้าร่วม'}
      </button>
    </div>
  );
}

function GroupsView({ cm, nav }) {
  const mine = GROUPS.filter((g) => cm.joined.includes(g.id));
  const others = GROUPS.filter((g) => !cm.joined.includes(g.id));
  return (
    <>
      <section>
        <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 10 }}>กลุ่มของฉัน</p>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {mine.length === 0
            ? <p style={{ fontSize: 12.5, color: 'var(--ink-soft)' }}>คุณยังไม่ได้เข้าร่วมกลุ่มใด เลือกกลุ่มที่สนใจจากรายการด้านล่าง</p>
            : mine.map((g) => <GroupRow key={g.id} group={g} cm={cm} nav={nav} />)}
        </div>
      </section>
      {others.length > 0 && (
        <section>
          <p className="font-display" style={{ fontSize: 15, fontWeight: 600, marginBottom: 10 }}>กลุ่มแนะนำ</p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {others.map((g) => <GroupRow key={g.id} group={g} cm={cm} nav={nav} />)}
          </div>
        </section>
      )}
    </>
  );
}

function CommunityScreen({ cm, nav }) {
  const tabs = [{ id: 'feed', label: 'ฟีด' }, { id: 'qa', label: 'ถาม-ตอบ' }, { id: 'groups', label: 'กลุ่ม' }];
  const tab = cm.ui.tab;
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 20px 0', flexShrink: 0 }}>
        <h1 className="font-display" style={{ fontSize: 20, fontWeight: 600, color: 'var(--ink)' }}>ชุมชน</h1>
        <button onClick={() => { cm.startPost({}); nav.push('create-post'); }} aria-label="สร้างโพสต์ใหม่"
          style={{ display: 'inline-flex', alignItems: 'center', gap: 5, padding: '9px 14px', borderRadius: 999, border: 'none', background: 'var(--pine)', color: '#fff', fontWeight: 700, fontSize: 12.5, cursor: 'pointer' }}>
          <Plus size={15} /> โพสต์
        </button>
      </div>
      <div role="tablist" aria-label="หมวดชุมชน" style={{ display: 'flex', borderBottom: '1px solid var(--border)', margin: '12px 20px 0', flexShrink: 0 }}>
        {tabs.map((t) => (
          <button key={t.id} role="tab" aria-selected={tab === t.id} onClick={() => cm.setUi({ tab: t.id })}
            style={{ flex: 1, padding: '10px 0', background: 'none', border: 'none', borderBottom: tab === t.id ? '2.5px solid var(--pine)' : '2.5px solid transparent', color: tab === t.id ? 'var(--pine)' : 'var(--ink-soft)', fontWeight: tab === t.id ? 700 : 500, fontSize: 13.5, cursor: 'pointer', marginBottom: -1 }}>
            {t.label}
          </button>
        ))}
      </div>
      <div className="scrollbar-hide" role="tabpanel" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: '14px 20px 20px', display: 'flex', flexDirection: 'column', gap: 14 }}>
        {tab === 'feed' && <FeedView cm={cm} nav={nav} />}
        {tab === 'qa' && <QaView cm={cm} nav={nav} />}
        {tab === 'groups' && <GroupsView cm={cm} nav={nav} />}
      </div>
    </div>
  );
}

function CommentItem({ comment, post, cm }) {
  const author = cm.memberOf(comment.authorId);
  const isPostAuthor = comment.authorId === post.authorId;
  const isLiked = cm.likedComments.includes(comment.id);
  const canMarkBest = post.type === 'question' && post.authorId === ME_ID && comment.authorId !== ME_ID;
  return (
    <div style={{ display: 'flex', gap: 10, padding: '12px 0', borderBottom: '1px solid var(--border)' }}>
      <Avatar member={author} size={32} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: 6 }}>
          <span style={{ fontSize: 12.5, fontWeight: 700, color: 'var(--ink)' }}>{author.name}</span>
          {isPostAuthor && <span style={{ fontSize: 10, fontWeight: 700, padding: '2px 8px', borderRadius: 999, background: 'var(--muted-bg)', color: 'var(--ink-soft)' }}>เจ้าของโพสต์</span>}
          {comment.best && <span style={{ display: 'inline-flex', alignItems: 'center', gap: 3, fontSize: 10, fontWeight: 700, padding: '2px 8px', borderRadius: 999, background: 'var(--pine-pale)', color: 'var(--pine)' }}><CheckCircle2 size={11} /> คำตอบที่ดีที่สุด</span>}
          <span style={{ fontSize: 10.5, color: 'var(--ink-soft)' }}>{comment.time}</span>
        </div>
        <p style={{ fontSize: 13, lineHeight: 1.6, color: 'var(--ink)', margin: '4px 0 6px', whiteSpace: 'pre-wrap' }}>{comment.text}</p>
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <button onClick={() => cm.toggleCommentLike(comment.id)} aria-pressed={isLiked} aria-label={isLiked ? 'ยกเลิกถูกใจความคิดเห็น' : 'ถูกใจความคิดเห็น'}
            style={{ display: 'inline-flex', alignItems: 'center', gap: 4, background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontSize: 11.5, fontWeight: 600, color: isLiked ? 'var(--danger)' : 'var(--ink-soft)' }}>
            <Heart size={14} fill={isLiked ? 'var(--danger)' : 'none'} /> {comment.likes + (isLiked ? 1 : 0)}
          </button>
          {canMarkBest && (
            <button onClick={() => cm.markBest(post.id, comment.id)}
              style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontSize: 11.5, fontWeight: 700, color: 'var(--pine)' }}>
              {comment.best ? 'ยกเลิกคำตอบที่ดีที่สุด' : 'เลือกเป็นคำตอบที่ดีที่สุด'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

function PostDetailScreen({ cm, nav }) {
  const [text, setText] = useState('');
  const post = cm.visiblePosts.find((p) => p.id === cm.selected.postId) || null;
  useEffect(() => { if (!post) nav.pop(); }, [post]);
  if (!post) return null;
  const list = [...cm.commentsOf(post.id)].sort((a, b) => (b.best ? 1 : 0) - (a.best ? 1 : 0));
  const submit = () => {
    if (!text.trim()) return;
    cm.addComment(post.id, text);
    setText('');
  };
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="โพสต์" onBack={nav.pop} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20 }}>
        <PostCard post={post} cm={cm} nav={nav} detail />
        <p className="font-display" style={{ fontSize: 15, fontWeight: 600, margin: '20px 0 2px' }}>
          {post.type === 'question' ? 'คำตอบ' : 'ความคิดเห็น'} ({list.length})
        </p>
        {list.length === 0
          ? <p style={{ fontSize: 12.5, color: 'var(--ink-soft)', padding: '16px 0' }}>{post.type === 'question' ? 'ยังไม่มีคำตอบ เป็นคนแรกที่ช่วยตอบคำถามนี้' : 'ยังไม่มีความคิดเห็น เริ่มบทสนทนาได้เลย'}</p>
          : list.map((c) => <CommentItem key={c.id} comment={c} post={post} cm={cm} />)}
      </div>
      <form onSubmit={(e) => { e.preventDefault(); submit(); }} style={{ display: 'flex', gap: 8, padding: '12px 20px', borderTop: '1px solid var(--border)', flexShrink: 0, background: 'var(--surface)' }}>
        <input value={text} onChange={(e) => setText(e.target.value)} maxLength={300} aria-label={post.type === 'question' ? 'เขียนคำตอบ' : 'เขียนความคิดเห็น'}
          placeholder={post.type === 'question' ? 'เขียนคำตอบของคุณ' : 'เขียนความคิดเห็น'}
          style={{ flex: 1, minWidth: 0, border: '1.5px solid var(--border)', borderRadius: 999, padding: '10px 16px', fontSize: 13.5, fontFamily: "'Plus Jakarta Sans', sans-serif", background: 'var(--paper)', color: 'var(--ink)', outline: 'none' }} />
        <button type="submit" disabled={!text.trim()} aria-label="ส่ง"
          style={{ width: 42, height: 42, borderRadius: 999, border: 'none', background: text.trim() ? 'var(--pine)' : 'var(--disabled)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: text.trim() ? 'pointer' : 'default', flexShrink: 0 }}>
          <Send size={17} color="#fff" />
        </button>
      </form>
    </div>
  );
}

function CreatePostScreen({ cm, nav }) {
  const c = cm.composer;
  const len = c.text.trim().length;
  const valid = len >= POST_MIN && len <= POST_MAX && !!c.destination;
  const myGroups = GROUPS.filter((g) => cm.joined.includes(g.id));
  const submit = () => { if (cm.createPost()) nav.pop(); };
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="สร้างโพสต์" onBack={nav.pop} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>ประเภทโพสต์</p>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            {COMPOSER_TYPES.map((t) => <Chip key={t} label={POST_TYPE_META[t].label} active={c.type === t} onClick={() => cm.updateComposer({ type: t })} />)}
          </div>
        </div>
        <div>
          <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>จุดหมายที่เกี่ยวข้อง</p>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            {[...DESTINATIONS.map((d) => d.name), 'ปาย'].filter((v, i, a) => a.indexOf(v) === i).map((d) => (
              <Chip key={d} label={d} active={c.destination === d} onClick={() => cm.updateComposer({ destination: d })} />
            ))}
          </div>
        </div>
        {myGroups.length > 0 && (
          <div>
            <p style={{ fontSize: 13, fontWeight: 700, marginBottom: 10 }}>โพสต์ในกลุ่ม (ไม่บังคับ)</p>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              <Chip label="ไม่ระบุกลุ่ม" active={!c.groupId} onClick={() => cm.updateComposer({ groupId: null })} />
              {myGroups.map((g) => <Chip key={g.id} label={g.name} active={c.groupId === g.id} onClick={() => cm.updateComposer({ groupId: g.id })} />)}
            </div>
          </div>
        )}
        <div>
          <label htmlFor="post-text" style={{ display: 'block', fontSize: 13, fontWeight: 700, marginBottom: 10 }}>เนื้อหา</label>
          <textarea id="post-text" rows={6} maxLength={POST_MAX} value={c.text} onChange={(e) => cm.updateComposer({ text: e.target.value })} placeholder={COMPOSER_HINT[c.type]}
            style={{ width: '100%', border: '1.5px solid var(--border)', borderRadius: 14, padding: '12px 14px', fontSize: 13.5, lineHeight: 1.6, fontFamily: "'Plus Jakarta Sans', sans-serif", color: 'var(--ink)', background: 'var(--surface)', outline: 'none', resize: 'none' }} />
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 6, fontSize: 11.5, color: 'var(--ink-soft)' }}>
            <span>{len > 0 && len < POST_MIN ? `พิมพ์อีก ${POST_MIN - len} ตัวอักษร` : ''}</span>
            <span>{c.text.length}/{POST_MAX}</span>
          </div>
        </div>
        <div style={{ borderRadius: 14, background: 'var(--pine-pale)', padding: 14 }}>
          <p style={{ fontSize: 12, fontWeight: 700, color: 'var(--pine)', marginBottom: 6 }}>แนวทางของชุมชน</p>
          {GROUP_RULES.map((r) => <p key={r} style={{ fontSize: 11.5, color: 'var(--ink-soft)', lineHeight: 1.6 }}>{r}</p>)}
        </div>
      </div>
      <div style={{ padding: '14px 20px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
        <PrimaryButton onClick={submit} disabled={!valid}>โพสต์</PrimaryButton>
        {!c.destination && <p style={{ fontSize: 11.5, color: 'var(--ink-soft)', textAlign: 'center', marginTop: 8 }}>เลือกจุดหมายที่เกี่ยวข้องก่อนโพสต์</p>}
      </div>
    </div>
  );
}

function MemberScreen({ cm, nav }) {
  const [tab, setTab] = useState('posts');
  const member = cm.memberOf(cm.selected.memberId || ME_ID);
  const isMe = member.id === ME_ID;
  const posts = cm.visiblePosts.filter((p) => p.authorId === member.id);
  const savedPosts = cm.visiblePosts.filter((p) => cm.saved.includes(p.id));
  const list = tab === 'saved' ? savedPosts : posts;
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title={isMe ? 'ชุมชนของฉัน' : 'โปรไฟล์'} onBack={nav.pop} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto', padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6, textAlign: 'center' }}>
          <Avatar member={member} size={72} />
          <p className="font-display" style={{ fontSize: 18, fontWeight: 600, color: 'var(--ink)' }}>{member.name}</p>
          <p style={{ fontSize: 12, fontWeight: 600, color: 'var(--pine)' }}>{member.level}</p>
          <p style={{ fontSize: 12.5, color: 'var(--ink-soft)', lineHeight: 1.6, maxWidth: 260 }}>{member.bio}</p>
        </div>
        <div style={{ display: 'flex', justifyContent: 'center', gap: 28 }}>
          <div style={{ textAlign: 'center' }}><p style={{ fontSize: 16, fontWeight: 700 }}>{posts.length}</p><p style={{ fontSize: 11, color: 'var(--ink-soft)' }}>โพสต์</p></div>
          <div style={{ textAlign: 'center' }}><p style={{ fontSize: 16, fontWeight: 700 }}>{formatCount(cm.followerCount(member))}</p><p style={{ fontSize: 11, color: 'var(--ink-soft)' }}>ผู้ติดตาม</p></div>
        </div>
        {!isMe && <div style={{ alignSelf: 'center' }}><FollowButton cm={cm} memberId={member.id} /></div>}
        {isMe && (
          <div style={{ display: 'flex', gap: 8, justifyContent: 'center' }}>
            <Chip label="โพสต์ของฉัน" active={tab === 'posts'} onClick={() => setTab('posts')} />
            <Chip label="ที่บันทึกไว้" active={tab === 'saved'} onClick={() => setTab('saved')} />
          </div>
        )}
        {list.length === 0
          ? <EmptyState icon={tab === 'saved' ? Bookmark : BookOpen} title={tab === 'saved' ? 'ยังไม่มีโพสต์ที่บันทึกไว้' : 'ยังไม่มีโพสต์'} hint={tab === 'saved' ? 'กดไอคอนบุ๊กมาร์กที่โพสต์เพื่อเก็บไว้อ่านทีหลัง' : isMe ? 'แบ่งปันประสบการณ์การเดินทางครั้งแรกของคุณ' : undefined}
              action={isMe && tab === 'posts' ? <div style={{ width: 200 }}><PrimaryButton onClick={() => { cm.startPost({}); nav.push('create-post'); }}>เขียนโพสต์</PrimaryButton></div> : undefined} />
          : list.map((p) => <PostCard key={p.id} post={p} cm={cm} nav={nav} linkAuthor={tab === 'saved' || p.authorId !== member.id} />)}
      </div>
    </div>
  );
}

function GroupDetailScreen({ cm, nav }) {
  const group = GROUPS.find((g) => g.id === cm.selected.groupId);
  if (!group) return null;
  const isJoined = cm.joined.includes(group.id);
  const list = cm.visiblePosts.filter((p) => p.groupId === group.id);
  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <TopBar title="กลุ่ม" onBack={nav.pop} />
      <div className="scrollbar-hide" style={{ flex: 1, minHeight: 0, overflowY: 'auto' }}>
        <div style={{ height: 120, background: group.gradient, display: 'flex', alignItems: 'flex-end', padding: 18 }}>
          <p className="font-display" style={{ color: '#fff', fontSize: 20, fontWeight: 600 }}>{group.name}</p>
        </div>
        <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
          <p style={{ fontSize: 13, lineHeight: 1.7, color: 'var(--ink)' }}>{group.desc}</p>
          <p style={{ fontSize: 12, color: 'var(--ink-soft)' }}>สมาชิก {cm.memberCount(group).toLocaleString()} คน</p>
          <div style={{ display: 'flex', gap: 10 }}>
            <div style={{ flex: 1 }}>
              {isJoined ? <SecondaryButton onClick={() => cm.toggleJoin(group.id)}>ออกจากกลุ่ม</SecondaryButton> : <PrimaryButton onClick={() => cm.toggleJoin(group.id)}>เข้าร่วมกลุ่ม</PrimaryButton>}
            </div>
            {isJoined && (
              <div style={{ flex: 1 }}><PrimaryButton onClick={() => { cm.startPost({ groupId: group.id }); nav.push('create-post'); }}>เขียนโพสต์</PrimaryButton></div>
            )}
          </div>
          <div style={{ borderRadius: 14, background: 'var(--pine-pale)', padding: 14 }}>
            <p style={{ fontSize: 12, fontWeight: 700, color: 'var(--pine)', marginBottom: 6 }}>กฎของกลุ่ม</p>
            {group.rules.map((r) => <p key={r} style={{ fontSize: 11.5, color: 'var(--ink-soft)', lineHeight: 1.6 }}>{r}</p>)}
          </div>
          <p className="font-display" style={{ fontSize: 15, fontWeight: 600 }}>โพสต์ในกลุ่ม</p>
          {list.length === 0
            ? <EmptyState icon={BookOpen} title="ยังไม่มีโพสต์ในกลุ่มนี้" hint={isJoined ? 'เริ่มบทสนทนาแรกของกลุ่ม' : 'เข้าร่วมกลุ่มเพื่อเริ่มโพสต์'} />
            : list.map((p) => <PostCard key={p.id} post={p} cm={cm} nav={nav} />)}
        </div>
      </div>
    </div>
  );
}

function SheetButton({ icon: Icon, label, onClick, danger }) {
  return (
    <button onClick={onClick} style={{ width: '100%', display: 'flex', alignItems: 'center', gap: 12, padding: '14px 4px', background: 'none', border: 'none', borderBottom: '1px solid var(--border)', cursor: 'pointer', textAlign: 'left', color: danger ? 'var(--danger)' : 'var(--ink)', fontSize: 14, fontWeight: 600 }}>
      {Icon && <Icon size={18} />} {label}
    </button>
  );
}

function CommunitySheet({ cm }) {
  const sheet = cm.sheet;
  useEffect(() => {
    if (!sheet) return undefined;
    const onKey = (e) => { if (e.key === 'Escape') cm.closeSheet(); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [sheet]);
  if (!sheet) return null;
  const post = cm.postById(sheet.postId);
  if (!post) return null;
  const isMine = post.authorId === ME_ID;
  const author = cm.memberOf(post.authorId);
  const close = cm.closeSheet;
  let title = 'ตัวเลือกโพสต์';
  let body;
  if (sheet.type === 'post') {
    const isSaved = cm.saved.includes(post.id);
    body = (
      <>
        <SheetButton icon={Bookmark} label={isSaved ? 'เลิกบันทึกโพสต์' : 'บันทึกโพสต์'} onClick={() => { cm.toggleSave(post.id); close(); }} />
        {!isMine && <SheetButton icon={cm.following.includes(author.id) ? UserCheck : UserPlus} label={cm.following.includes(author.id) ? `เลิกติดตาม ${author.name}` : `ติดตาม ${author.name}`} onClick={() => { cm.toggleFollow(author.id); close(); }} />}
        {!isMine && <SheetButton icon={EyeOff} label="ซ่อนโพสต์นี้" onClick={() => { cm.hidePost(post.id); close(); }} />}
        {!isMine && <SheetButton icon={Flag} label="รายงานโพสต์" danger onClick={() => cm.openSheet({ type: 'report', postId: post.id })} />}
        {isMine && <SheetButton icon={Trash2} label="ลบโพสต์" danger onClick={() => cm.openSheet({ type: 'delete', postId: post.id })} />}
      </>
    );
  } else if (sheet.type === 'report') {
    title = 'รายงานโพสต์';
    body = (
      <>
        <p style={{ fontSize: 12, color: 'var(--ink-soft)', marginBottom: 4 }}>เลือกเหตุผลที่ตรงที่สุด ทีมงานจะตรวจสอบและโพสต์นี้จะถูกซ่อนจากฟีดของคุณ</p>
        {REPORT_REASONS.map((r) => <SheetButton key={r} label={r} onClick={() => { cm.reportPost(post.id); close(); }} />)}
      </>
    );
  } else {
    title = 'ลบโพสต์นี้?';
    body = (
      <>
        <p style={{ fontSize: 12.5, color: 'var(--ink-soft)', lineHeight: 1.6, marginBottom: 14 }}>โพสต์และความคิดเห็นทั้งหมดจะถูกลบและไม่สามารถกู้คืนได้</p>
        <div style={{ display: 'flex', gap: 10 }}>
          <div style={{ flex: 1 }}><SecondaryButton onClick={close}>ยกเลิก</SecondaryButton></div>
          <button onClick={() => { cm.deletePost(post.id); close(); }} style={{ flex: 1, padding: 15, borderRadius: 16, border: 'none', background: 'var(--danger)', color: '#fff', fontWeight: 700, fontSize: 15, cursor: 'pointer' }}>ลบโพสต์</button>
        </div>
      </>
    );
  }
  return (
    <div style={{ position: 'absolute', inset: 0, zIndex: 60, display: 'flex', flexDirection: 'column', justifyContent: 'flex-end' }}>
      <div onClick={close} aria-hidden="true" style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.45)' }} />
      <div role="dialog" aria-modal="true" aria-label={title} className="fade-up" style={{ position: 'relative', background: 'var(--surface)', borderTopLeftRadius: 24, borderTopRightRadius: 24, padding: '16px 20px 28px' }}>
        <div style={{ width: 40, height: 4, borderRadius: 999, background: 'var(--border)', margin: '0 auto 12px' }} />
        <p className="font-display" style={{ fontSize: 16, fontWeight: 600, marginBottom: 6 }}>{title}</p>
        {body}
      </div>
    </div>
  );
}

/* ---------------------------------------------------------------- */
/*  MAIN APP                                                          */
/* ---------------------------------------------------------------- */

export default function SabaiTravelApp() {
  const [stack, setStack] = useState(['login']);
  const screen = stack[stack.length - 1];
  const push = (s) => setStack((prev) => [...prev, s]);
  const pop = () => setStack((prev) => (prev.length > 1 ? prev.slice(0, -1) : prev));
  const gotoTab = (tab) => setStack([tab]);

  const [loginEmail, setLoginEmail] = useState('');
  const [loginPassword, setLoginPassword] = useState('');

  const [searchQuery, setSearchQuery] = useState('');
  const [searchCategory, setSearchCategory] = useState('ทั้งหมด');
  const [selectedListing, setSelectedListing] = useState(null);

  const [aiDestination, setAiDestination] = useState('เชียงใหม่');
  const [aiDays, setAiDays] = useState(3);
  const [aiBudget, setAiBudget] = useState(10000);
  const [aiStyles, setAiStyles] = useState(['ธรรมชาติ', 'อาหารท้องถิ่น']);
  const [aiTravelers, setAiTravelers] = useState(2);
  const [aiGenerating, setAiGenerating] = useState(false);
  const [aiResult, setAiResult] = useState(null);
  const [aiSaved, setAiSaved] = useState(false);
  const [savedTrips, setSavedTrips] = useState([]);

  const [bookingItem, setBookingItem] = useState(null);
  const [paymentMethod, setPaymentMethod] = useState('promptpay');
  const [processingPayment, setProcessingPayment] = useState(false);
  const [bookingNumber, setBookingNumber] = useState('');

  const [registerForm, setRegisterForm] = useState({
    firstName: '', lastName: '', email: '', phone: '', password: '', confirmPassword: '',
    gender: 'หญิง', dob: '', nationality: 'ไทย', language: 'ไทย',
    budgetLevel: 'ปานกลาง', travelStyles: [], favoriteDestination: '', travelWith: 'คู่รัก',
  });
  const updateRegisterForm = (patch) => setRegisterForm((prev) => ({ ...prev, ...patch }));
  const toggleRegisterStyle = (s) => setRegisterForm((prev) => ({
    ...prev,
    travelStyles: prev.travelStyles.includes(s) ? prev.travelStyles.filter((x) => x !== s) : [...prev.travelStyles, s],
  }));
  const handleFinishRegister = () => gotoTab('home');

  const [walletBalance, setWalletBalance] = useState(4250);
  const [walletTransactions, setWalletTransactions] = useState([
    { id: 1, type: 'in', label: 'เติมเงินผ่านพร้อมเพย์', date: '5 ก.ย. 2569', amount: 2000 },
    { id: 2, type: 'out', label: 'ชำระค่าที่พัก ล้านนา วิว รีสอร์ท', date: '2 ก.ย. 2569', amount: 1490 },
    { id: 3, type: 'in', label: 'เงินคืนจากการยกเลิกจอง', date: '28 ส.ค. 2569', amount: 990 },
    { id: 4, type: 'out', label: 'ชำระค่าอาหาร ข้าวซอยแม่สาย', date: '20 ส.ค. 2569', amount: 120 },
  ]);
  const handleTopUp = () => {
    setWalletBalance((prev) => prev + 1000);
    setWalletTransactions((prev) => [{ id: Date.now(), type: 'in', label: 'เติมเงินผ่านพร้อมเพย์', date: 'วันนี้', amount: 1000 }, ...prev]);
  };

  const [selectedTrip, setSelectedTrip] = useState(null);

  const [notifications, setNotifications] = useState([
    { id: 1, type: 'booking', title: 'ยืนยันการจองสำเร็จ', message: 'การจองที่ ล้านนา วิว รีสอร์ท ของคุณได้รับการยืนยันแล้ว', time: '5 นาทีที่แล้ว', read: false },
    { id: 2, type: 'payment', title: 'ชำระเงินสำเร็จ', message: 'ชำระเงิน ฿1,564 ผ่านพร้อมเพย์เรียบร้อยแล้ว', time: '5 นาทีที่แล้ว', read: false },
    { id: 3, type: 'promotion', title: 'โปรโมชั่นใหม่! ลด 20%', message: 'ใช้โค้ด SABAI20 สำหรับทริปแรกของคุณ ก่อนสิ้นเดือนนี้', time: '2 ชั่วโมงที่แล้ว', read: false },
    { id: 4, type: 'trip', title: 'ทริปของคุณใกล้ถึงแล้ว', message: 'อีก 3 วัน ทริปเชียงใหม่ของคุณจะเริ่มต้น เตรียมตัวให้พร้อม', time: 'เมื่อวาน', read: true },
    { id: 5, type: 'system', title: 'อัปเดตนโยบายความเป็นส่วนตัว', message: 'เราปรับปรุงนโยบายความเป็นส่วนตัวใหม่ อ่านรายละเอียดเพิ่มเติมได้ในการตั้งค่า', time: '3 วันที่แล้ว', read: true },
  ]);
  const unreadCount = notifications.filter((n) => !n.read).length;
  const handleMarkAllRead = () => setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
  const handleMarkOneRead = (id) => setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, read: true } : n)));

  const [reviews, setReviews] = useState([
    { id: 1, itemName: 'ล้านนา วิว รีสอร์ท', itemType: 'hotel', rating: 5, comment: 'ที่พักสวยมาก บรรยากาศดี พนักงานเป็นกันเอง จะกลับมาพักอีกแน่นอน', date: '28 ส.ค. 2569' },
    { id: 2, itemName: 'ข้าวซอยแม่สาย', itemType: 'restaurant', rating: 4, comment: 'อร่อยมาก แต่คนเยอะไปหน่อยช่วงเที่ยง ควรไปก่อน 11 โมง', date: '15 ส.ค. 2569' },
  ]);
  const [writeReviewItem, setWriteReviewItem] = useState(null);
  const [writeReviewRating, setWriteReviewRating] = useState(0);
  const [writeReviewComment, setWriteReviewComment] = useState('');
  const handleSubmitReview = () => {
    setReviews((prev) => [{ id: Date.now(), itemName: writeReviewItem.name, itemType: writeReviewItem.type, rating: writeReviewRating, comment: writeReviewComment, date: 'วันนี้' }, ...prev]);
    setWriteReviewItem(null);
    setWriteReviewRating(0);
    setWriteReviewComment('');
    pop();
  };

  const [coupons, setCoupons] = useState(INITIAL_COUPONS);
  const handleApplyCoupon = (coupon) => {
    setCoupons((prev) => prev.map((c) => (c.id === coupon.id ? { ...c, status: 'used' } : c)));
  };

  const [supportTickets, setSupportTickets] = useState([]);
  const [ticketSubject, setTicketSubject] = useState('');
  const [ticketMessage, setTicketMessage] = useState('');
  const handleSubmitTicket = () => {
    setSupportTickets((prev) => [{ id: Date.now(), subject: ticketSubject, message: ticketMessage, date: 'วันนี้' }, ...prev]);
    setTicketSubject('');
    setTicketMessage('');
    pop();
  };

  const [bookings, setBookings] = useState([
    { id: 'b3', number: 'SB495530', type: 'flight', title: 'กรุงเทพฯ → เชียงใหม่', subtitle: 'Sabai Air · เดินทาง 25 ต.ค. 2569', amount: 1040, method: 'บัตรเครดิต/เดบิต', date: '10 ต.ค. 2569', status: 'confirmed' },
    { id: 'b1', number: 'SB482913', type: 'hotel', title: 'ล้านนา วิว รีสอร์ท', subtitle: 'นิมมานเหมินท์, เชียงใหม่', amount: 1564, method: 'พร้อมเพย์ (PromptPay)', date: '2 ก.ย. 2569', status: 'completed' },
    { id: 'b2', number: 'SB471225', type: 'restaurant', title: 'ข้าวซอยแม่สาย', subtitle: 'เมืองเก่า, เชียงใหม่', amount: 126, method: 'Sabai Wallet', date: '20 ส.ค. 2569', status: 'completed' },
  ]);
  const [selectedBookingId, setSelectedBookingId] = useState(null);
  const selectedBooking = bookings.find((b) => b.id === selectedBookingId) || null;
  const [cancelReason, setCancelReason] = useState('');
  const [paymentError, setPaymentError] = useState('');

  const [settings, setSettings] = useState({ language: 'ไทย', currency: 'THB ฿', theme: 'สว่าง', notif: { push: true, email: true, trip: true, promo: false } });
  const [systemDark, setSystemDark] = useState(false);
  useEffect(() => {
    try {
      const mq = window.matchMedia('(prefers-color-scheme: dark)');
      setSystemDark(mq.matches);
      const onChange = (e) => setSystemDark(e.matches);
      mq.addEventListener('change', onChange);
      return () => mq.removeEventListener('change', onChange);
    } catch (e) { return undefined; }
  }, []);
  const isDark = settings.theme === 'มืด' || (settings.theme === 'ตามระบบ' && systemDark);

  const [toast, setToast] = useState(null);
  const toastTimer = useRef(null);
  const showToast = (message) => {
    setToast(message);
    clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => setToast(null), 2400);
  };
  const pushNotification = (type, title, message) =>
    setNotifications((prev) => [{ id: Date.now() + Math.random(), type, title, message, time: 'เมื่อสักครู่', read: false }, ...prev]);

  const cm = useCommunity({
    showToast, pushNotification,
    onSaveTrip: (post) => setSavedTrips((prev) => [...prev, {
      id: Date.now(), destination: post.destination, days: post.sharedTrip.days, budget: post.sharedTrip.budget, travelers: 2,
      itinerary: buildItinerary({ destinationName: post.destination, days: post.sharedTrip.days, budget: post.sharedTrip.budget, styles: post.sharedTrip.styles }),
    }]),
  });
  const nav = { push, pop };

  const handleSettingsChange = (patch) => {
    setSettings((prev) => ({ ...prev, ...patch }));
    if (!patch.notif && !patch.theme) showToast('บันทึกการตั้งค่าแล้ว');
  };
  const handleChangePassword = () => { showToast('เปลี่ยนรหัสผ่านเรียบร้อยแล้ว'); pop(); };
  const handleDeleteAccount = () => {
    setStack(['login']);
    setSavedTrips([]);
    showToast('ลบบัญชีเรียบร้อยแล้ว');
  };

  const handleSelectBooking = (id) => { setSelectedBookingId(id); push('booking-detail'); };
  const handleReviewBooking = () => {
    const match = LISTINGS.find((l) => l.name === selectedBooking.title);
    if (match) setWriteReviewItem(match);
    push('write-review');
  };
  const handleConfirmCancel = () => {
    const b = selectedBooking;
    const refund = refundFor(b);
    setBookings((prev) => prev.map((x) => (x.id === b.id ? { ...x, status: 'cancelled', cancelReason, refund } : x)));
    setWalletBalance((prev) => prev + refund);
    setWalletTransactions((prev) => [{ id: Date.now(), type: 'in', label: `เงินคืนจากการยกเลิก ${b.title}`, date: 'วันนี้', amount: refund }, ...prev]);
    pushNotification('booking', 'ยกเลิกการจองแล้ว', `${b.title} (${b.number}) ถูกยกเลิก เงินคืน ฿${refund.toLocaleString()} เข้า Sabai Wallet แล้ว`);
    setCancelReason('');
    showToast(`ยกเลิกแล้ว คืนเงิน ฿${refund.toLocaleString()} เข้า Wallet`);
    pop();
  };

  const toggleAiStyle = (s) => setAiStyles((prev) => (prev.includes(s) ? prev.filter((x) => x !== s) : [...prev, s]));

  const handleGenerateAi = () => {
    setAiGenerating(true);
    setAiSaved(false);
    setTimeout(() => {
      setAiResult(buildItinerary({ destinationName: aiDestination, days: aiDays, budget: aiBudget, styles: aiStyles }));
      setAiGenerating(false);
    }, 1700);
  };
  const handleResetAi = () => setAiResult(null);
  const handleSaveAiTrip = () => {
    setAiSaved(true);
    const total = aiResult.budgetBreakdown.reduce((s, b) => s + b.value, 0);
    setSavedTrips((prev) => [...prev, {
      id: Date.now(), destination: aiDestination, days: aiDays, budget: total,
      travelers: aiTravelers, itinerary: aiResult,
    }]);
  };
  const handleBookAiTrip = () => {
    const total = aiResult.budgetBreakdown.reduce((s, b) => s + b.value, 0);
    setBookingItem({ title: `ทริป ${aiDestination} ${aiDays} วัน`, subtitle: `${aiTravelers} ผู้เดินทาง · จัดโดย AI`, price: total, type: 'trip' });
    push('booking');
  };

  const filteredListings = LISTINGS.filter((l) => {
    const catMatch = searchCategory === 'ทั้งหมด' || TYPE_LABEL[l.type] === searchCategory;
    const qMatch = searchQuery === '' || l.name.includes(searchQuery) || l.location.includes(searchQuery);
    return catMatch && qMatch;
  });

  const handleSelectListing = (item) => { setSelectedListing(item); push('listing-detail'); };
  const handleBookListing = () => {
    setBookingItem({ title: selectedListing.name, subtitle: selectedListing.location, price: selectedListing.price, type: selectedListing.type });
    push('booking');
  };
  const handleQuickAction = (label) => { setSearchCategory(label); setSearchQuery(''); push('search'); };
  const handleDestinationTap = (d) => { setSearchQuery(d.name); setSearchCategory('ทั้งหมด'); push('search'); };

  const PAYMENT_LABEL = { promptpay: 'พร้อมเพย์ (PromptPay)', card: 'บัตรเครดิต/เดบิต', wallet: 'Sabai Wallet' };
  const handleConfirmPayment = () => {
    const total = Math.round(bookingItem.price * 1.05);
    if (paymentMethod === 'wallet' && total > walletBalance) {
      setPaymentError(`ยอด Sabai Wallet ไม่เพียงพอ (คงเหลือ ฿${walletBalance.toLocaleString()}) กรุณาเติมเงินหรือเลือกวิธีชำระอื่น`);
      return;
    }
    setPaymentError('');
    setProcessingPayment(true);
    setTimeout(() => {
      const number = 'SB' + Math.floor(100000 + Math.random() * 900000);
      setProcessingPayment(false);
      setBookingNumber(number);
      setBookings((prev) => [{ id: 'b' + Date.now(), number, type: bookingItem.type, title: bookingItem.title, subtitle: bookingItem.subtitle, amount: total, method: PAYMENT_LABEL[paymentMethod], date: 'วันนี้', status: 'confirmed' }, ...prev]);
      if (paymentMethod === 'wallet') {
        setWalletBalance((prev) => prev - total);
        setWalletTransactions((prev) => [{ id: Date.now(), type: 'out', label: `ชำระค่าจอง ${bookingItem.title}`, date: 'วันนี้', amount: total }, ...prev]);
      }
      pushNotification('booking', 'ยืนยันการจองสำเร็จ', `${bookingItem.title} หมายเลข ${number} ได้รับการยืนยันแล้ว`);
      pushNotification('payment', 'ชำระเงินสำเร็จ', `ชำระเงิน ฿${total.toLocaleString()} ผ่าน ${PAYMENT_LABEL[paymentMethod]} เรียบร้อยแล้ว`);
      push('confirmation');
    }, 1600);
  };

  const handleBackToHome = () => {
    setStack(['home']);
    setAiResult(null);
    setAiSaved(false);
    setBookingItem(null);
  };

  const showTabBar = ['home', 'search', 'ai', 'trips', 'community', 'profile'].includes(screen);

  let content = null;
  if (screen === 'login') {
    content = <LoginScreen email={loginEmail} setEmail={setLoginEmail} password={loginPassword} setPassword={setLoginPassword} onLogin={() => gotoTab('home')} onGuest={() => gotoTab('home')} onRegister={() => push('register-1')} />;
  } else if (screen === 'register-1') {
    content = <RegisterScreen step={1} form={registerForm} onChange={updateRegisterForm} onToggleStyle={toggleRegisterStyle} onBack={pop} onNext={() => push('register-2')} />;
  } else if (screen === 'register-2') {
    content = <RegisterScreen step={2} form={registerForm} onChange={updateRegisterForm} onToggleStyle={toggleRegisterStyle} onBack={pop} onNext={() => push('register-3')} />;
  } else if (screen === 'register-3') {
    content = <RegisterScreen step={3} form={registerForm} onChange={updateRegisterForm} onToggleStyle={toggleRegisterStyle} onBack={pop} onFinish={handleFinishRegister} />;
  } else if (screen === 'wallet') {
    content = <WalletScreen balance={walletBalance} transactions={walletTransactions} onBack={pop} onTopUp={handleTopUp} />;
  } else if (screen === 'trip-detail') {
    content = <TripDetailScreen trip={selectedTrip} onBack={pop} />;
  } else if (screen === 'notifications') {
    content = <NotificationScreen notifications={notifications} onBack={pop} onMarkAll={handleMarkAllRead} onMarkOne={handleMarkOneRead} />;
  } else if (screen === 'reviews') {
    content = <ReviewsScreen reviews={reviews} onBack={pop} onWriteReview={() => push('write-review')} />;
  } else if (screen === 'write-review') {
    content = (
      <WriteReviewScreen
        selectedItem={writeReviewItem} onSelectItem={setWriteReviewItem}
        rating={writeReviewRating} setRating={setWriteReviewRating}
        comment={writeReviewComment} setComment={setWriteReviewComment}
        onBack={pop} onSubmit={handleSubmitReview}
      />
    );
  } else if (screen === 'bookings') {
    content = <BookingsScreen bookings={bookings} onBack={pop} onSelect={handleSelectBooking} onStartSearch={() => { setSearchCategory('ทั้งหมด'); setSearchQuery(''); gotoTab('search'); }} />;
  } else if (screen === 'booking-detail') {
    content = <BookingDetailScreen booking={selectedBooking} onBack={pop} onCancel={() => push('cancel-booking')} onReview={selectedBooking && LISTINGS.some((l) => l.name === selectedBooking.title) ? handleReviewBooking : undefined} />;
  } else if (screen === 'cancel-booking') {
    content = <CancelBookingScreen booking={selectedBooking} reason={cancelReason} setReason={setCancelReason} onBack={pop} onConfirm={handleConfirmCancel} />;
  } else if (screen === 'settings') {
    content = <SettingsScreen settings={settings} onChange={handleSettingsChange} onChangePassword={() => push('change-password')} onDeleteAccount={handleDeleteAccount} onBack={pop} />;
  } else if (screen === 'change-password') {
    content = <ChangePasswordScreen onBack={pop} onSubmit={handleChangePassword} />;
  } else if (screen === 'coupons') {
    content = <CouponScreen coupons={coupons} onBack={pop} onApply={handleApplyCoupon} />;
  } else if (screen === 'help') {
    content = <HelpScreen onBack={pop} onOpenTicket={() => push('support-ticket')} tickets={supportTickets} />;
  } else if (screen === 'support-ticket') {
    content = (
      <SupportTicketScreen
        subject={ticketSubject} setSubject={setTicketSubject}
        message={ticketMessage} setMessage={setTicketMessage}
        onBack={pop} onSubmit={handleSubmitTicket}
      />
    );
  } else if (screen === 'home') {
    content = <HomeScreen onSearchTap={() => { setSearchCategory('ทั้งหมด'); setSearchQuery(''); push('search'); }} onQuickAction={handleQuickAction} onAiTap={() => push('ai')} onDestinationTap={handleDestinationTap} onNotificationsTap={() => push('notifications')} unreadCount={unreadCount} />;
  } else if (screen === 'search') {
    content = <SearchScreen query={searchQuery} setQuery={setSearchQuery} category={searchCategory} setCategory={setSearchCategory} results={filteredListings} onSelect={handleSelectListing} />;
  } else if (screen === 'listing-detail') {
    content = <ListingDetailScreen item={selectedListing} onBack={pop} onBook={handleBookListing} />;
  } else if (screen === 'ai') {
    content = (
      <AiPlannerScreen
        destination={aiDestination} setDestination={setAiDestination}
        days={aiDays} setDays={setAiDays}
        budget={aiBudget} setBudget={setAiBudget}
        styles={aiStyles} toggleStyle={toggleAiStyle}
        travelers={aiTravelers} setTravelers={setAiTravelers}
        generating={aiGenerating} result={aiResult}
        onGenerate={handleGenerateAi} onReset={handleResetAi}
        onSave={handleSaveAiTrip} onBookTrip={handleBookAiTrip} saved={aiSaved}
      />
    );
  } else if (screen === 'trips') {
    content = <TripsScreen savedTrips={savedTrips} onStartAi={() => push('ai')} onSelectTrip={(t) => { setSelectedTrip(t); push('trip-detail'); }} />;
  } else if (screen === 'community') {
    content = <CommunityScreen cm={cm} nav={nav} />;
  } else if (screen === 'post-detail') {
    content = <PostDetailScreen cm={cm} nav={nav} />;
  } else if (screen === 'create-post') {
    content = <CreatePostScreen cm={cm} nav={nav} />;
  } else if (screen === 'member') {
    content = <MemberScreen cm={cm} nav={nav} />;
  } else if (screen === 'group-detail') {
    content = <GroupDetailScreen cm={cm} nav={nav} />;
  } else if (screen === 'profile') {
    content = <ProfileScreen onCommunityTap={() => { cm.selectMember('me'); push('member'); }} onLogout={() => setStack(['login'])} onWalletTap={() => push('wallet')} onReviewsTap={() => push('reviews')} onCouponTap={() => push('coupons')} onHelpTap={() => push('help')} onBookingsTap={() => push('bookings')} onSettingsTap={() => push('settings')} />;
  } else if (screen === 'booking') {
    content = <BookingScreen item={bookingItem} onBack={pop} onPay={() => push('payment')} />;
  } else if (screen === 'payment') {
    content = <PaymentScreen total={bookingItem ? Math.round(bookingItem.price * 1.05) : 0} method={paymentMethod} setMethod={(m) => { setPaymentMethod(m); setPaymentError(''); }} error={paymentError} onBack={pop} onConfirm={handleConfirmPayment} processing={processingPayment} walletBalance={walletBalance} />;
  } else if (screen === 'confirmation') {
    content = <ConfirmationScreen bookingNumber={bookingNumber} item={bookingItem} onHome={handleBackToHome} onViewBookings={() => { setStack(['profile', 'bookings']); setBookingItem(null); }} />;
  }

  return (
    <div style={{ minHeight: '700px', width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'radial-gradient(circle at 25% 15%, #1C5C57 0%, #0B2422 55%, #071614 100%)', padding: '28px 14px', fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
      <style>{FONT_CSS}</style>
      <div className={isDark ? 'theme-dark' : ''} style={{ width: '100%', maxWidth: 390, height: 'min(830px, 92vh)', background: 'var(--paper)', borderRadius: 40, boxShadow: '0 40px 70px -25px rgba(0,0,0,0.55), 0 0 0 9px #0B2422', overflow: 'hidden', display: 'flex', flexDirection: 'column', position: 'relative' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 24px 4px', fontSize: 12, fontWeight: 700, color: 'var(--ink)', flexShrink: 0 }}>
          <span>9:41</span>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <Wifi size={13} />
            <div style={{ width: 22, height: 11, border: '1.4px solid var(--ink)', borderRadius: 3, position: 'relative', display: 'flex', alignItems: 'center', padding: 1 }}>
              <div style={{ width: '75%', height: '100%', background: 'var(--ink)', borderRadius: 1 }} />
            </div>
          </div>
        </div>
        <div style={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
          {content}
        </div>
        {showTabBar && <TabBar active={screen} onChange={gotoTab} />}
        <CommunitySheet cm={cm} />
        <Toast message={toast} />
      </div>
    </div>
  );
}
