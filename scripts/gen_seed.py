#!/usr/bin/env python3
"""Single source of truth for DEV seed/fixture data.

Generates (keeps in sync):
  services/api-go/migrations/0002_seed_dev.up.sql
  services/api-go/internal/api/fixtures.go
  apps/mobile-flutter/lib/data/mock_fixtures.dart
Usage: python3 scripts/gen_seed.py   (run from repo root; use --check to fail on drift)
"""
import json, sys, pathlib

ROOT = pathlib.Path(__file__).resolve().parent.parent
D = [(1,'phuket','ภูเก็ต','หาดทราย พระอาทิตย์ตก','linear-gradient(135deg,#0b5ed7,#2bb3c0)'),
(2,'chiang-mai','เชียงใหม่','วัฒนธรรมล้านนา อากาศเย็น','linear-gradient(135deg,#1b7f5c,#8bc34a)'),
(3,'krabi','กระบี่','หน้าผาและทะเลอันดามัน','linear-gradient(135deg,#0f766e,#38bdf8)'),
(4,'bangkok','กรุงเทพฯ','เมืองที่ไม่เคยหลับ','linear-gradient(135deg,#002765,#6366f1)'),
(5,'samui','เกาะสมุย','พักผ่อนริมอ่าวไทย','linear-gradient(135deg,#f59e0b,#ef4444)'),
(6,'pai','ปาย','หมอกเช้าและคาเฟ่ริมน้ำ','linear-gradient(135deg,#7c3aed,#22c55e)')]
L = [(1,1,'hotel','Andaman Sunset Resort','หาดป่าตอง ภูเก็ต',4.8,1240,3200,['สระว่ายน้ำ','อาหารเช้า','Wi-Fi','ใกล้ชายหาด']),
(2,1,'hotel','Kata Beach Boutique','หาดกะตะ ภูเก็ต',4.6,860,2100,['Wi-Fi','อาหารเช้า','รถรับส่ง']),
(3,1,'restaurant','ครัวทะเลเก่าภูเก็ต','เมืองเก่า ภูเก็ต',4.7,532,450,['อาหารทะเลสด','รับจองโต๊ะ']),
(4,1,'attraction','ทัวร์เกาะพีพี','ท่าเรือรัษฎา ภูเก็ต',4.9,2100,1490,['เรือสปีดโบ๊ท','อาหารกลางวัน','อุปกรณ์ดำน้ำ']),
(5,2,'hotel','Lanna Courtyard Hotel','นิมมานเหมินทร์ เชียงใหม่',4.7,640,1800,['Wi-Fi','สปา','อาหารเช้า']),
(6,2,'restaurant','ข้าวซอยลำดวน','ช้างคลาน เชียงใหม่',4.5,980,120,['อาหารเหนือ','ราคาประหยัด']),
(7,2,'attraction','ดอยสุเทพ + ไนท์ซาฟารี','เชียงใหม่',4.6,415,990,['รถรับส่ง','ไกด์']),
(8,3,'hotel','Railay Cliff Villas','ไร่เลย์ กระบี่',4.8,370,4200,['สระส่วนตัว','วิวทะเล','อาหารเช้า']),
(9,4,'flight','BKK → HKT บินตรง','สุวรรณภูมิ → ภูเก็ต',4.4,300,1890,['ตรงเวลา','น้ำหนักกระเป๋า 20 กก.']),
(10,4,'hotel','Riverside Bangkok Hotel','ริมแม่น้ำเจ้าพระยา กรุงเทพฯ',4.6,1500,2600,['สระว่ายน้ำ','Wi-Fi','ใกล้ BTS']),
(11,5,'trip','แพ็กเกจเกาะสมุย 3 วัน 2 คืน','เกาะสมุย',4.7,210,7900,['โรงแรม 2 คืน','รถรับส่ง','ทัวร์เกาะ']),
(12,4,'restaurant','เจ๊ไฝ สตรีทฟู้ด','ประตูผี กรุงเทพฯ',4.8,3300,650,['มิชลินไกด์','ปูผัดผงกะหรี่'])]
F = [('จองแล้วยกเลิกได้หรือไม่?','ยกเลิกได้ที่เมนู การจองของฉัน > รายละเอียด > ยกเลิกการจอง เงินคืนเข้า Wallet ตามเงื่อนไขของรายการ'),
('ชำระเงินด้วยวิธีใดได้บ้าง?','รองรับ PromptPay บัตรเครดิต/เดบิต และ Wallet'),
('Wallet เติมเงินอย่างไร?','ไปที่ Wallet > เติมเงิน แล้วสแกน QR PromptPay'),
('ติดต่อฝ่ายช่วยเหลือได้อย่างไร?','เปิดหน้า ติดต่อเรา แล้วส่งข้อความ ทีมงานตอบกลับภายใน 24 ชั่วโมง')]

# ---- demo catalog extension (migration 0006, only with MIGRATE_SEED) -------------------------------------------
D0, L0, F0 = D, L, F
D2 = [(7,'ayutthaya','อยุธยา','มรดกโลกและวัดเก่าแก่','linear-gradient(135deg,#92400e,#f59e0b)'),
(8,'khao-yai','เขาใหญ่','ธรรมชาติ ไวน์ และอากาศเย็น','linear-gradient(135deg,#166534,#84cc16)')]
L2 = [(13,1,'hotel','Patong Bay Residence','หาดป่าตอง ภูเก็ต',4.5,720,2400,['สระว่ายน้ำ','Wi-Fi','ฟิตเนส']),
(14,1,'restaurant','โอเชียนเทอเรซ ซีฟู้ด','หาดกะรน ภูเก็ต',4.6,410,890,['วิวทะเล','อาหารทะเลสด','รับจองโต๊ะ']),
(15,1,'attraction','ดำน้ำตื้นเกาะราชา','ท่าเรือฉลอง ภูเก็ต',4.7,640,1650,['อุปกรณ์ดำน้ำ','อาหารกลางวัน','ไกด์']),
(16,2,'hotel','Doi Suthep Hillside Lodge','สุเทพ เชียงใหม่',4.6,305,2200,['วิวภูเขา','อาหารเช้า','Wi-Fi']),
(17,2,'restaurant','เฮือนเจ้านาง','ในเมือง เชียงใหม่',4.7,520,380,['อาหารล้านนา','บรรยากาศดั้งเดิม']),
(18,2,'attraction','ปางช้างจริยธรรม','แม่แตง เชียงใหม่',4.9,1180,2300,['ไม่ขี่ช้าง','อาหารกลางวัน','รถรับส่ง']),
(19,3,'attraction','ทัวร์ 4 เกาะ ทะเลแหวก','อ่าวนาง กระบี่',4.7,890,1100,['เรือหางยาว','อุปกรณ์ดำน้ำ','อาหารกลางวัน']),
(20,3,'restaurant','ครัวอ่าวนาง','อ่าวนาง กระบี่',4.4,260,520,['อาหารใต้','ราคาประหยัด']),
(21,4,'hotel','Sukhumvit Urban Suites','สุขุมวิท กรุงเทพฯ',4.5,1320,2900,['ใกล้ BTS','สระว่ายน้ำ','Co-working']),
(22,4,'attraction','ล่องเรือเจ้าพระยาพร้อมมื้อค่ำ','ท่าเรือ ICONSIAM กรุงเทพฯ',4.6,1750,1290,['บุฟเฟต์','วิวพระบรมมหาราชวัง']),
(23,4,'flight','BKK → CNX บินตรง','สุวรรณภูมิ → เชียงใหม่',4.5,410,1450,['ตรงเวลา','น้ำหนักกระเป๋า 20 กก.']),
(24,4,'flight','DMK → KBV บินตรง','ดอนเมือง → กระบี่',4.3,230,1290,['ราคาประหยัด','อาหารบนเครื่อง']),
(25,5,'hotel','Chaweng Palm Resort','หาดเฉวง เกาะสมุย',4.6,680,3600,['ติดชายหาด','สระว่ายน้ำ','สปา']),
(26,5,'attraction','ทัวร์อ่างทอง 1 วัน','เกาะสมุย',4.8,540,1850,['เรือสปีดโบ๊ท','พายคายัค','อาหารกลางวัน']),
(27,6,'hotel','Pai Riverside Bungalows','ริมน้ำปาย แม่ฮ่องสอน',4.7,330,1500,['ริมแม่น้ำ','อาหารเช้า','Wi-Fi']),
(28,6,'trip','แพ็กเกจปาย 3 วัน 2 คืน','ปาย แม่ฮ่องสอน',4.6,145,5600,['ที่พัก 2 คืน','รถตู้','ทัวร์น้ำตก']),
(29,7,'trip','แพ็กเกจอยุธยา วันเดียว','อยุธยา',4.5,260,1250,['รถรับส่ง','ไกด์','ค่าเข้าชม']),
(30,8,'hotel','Khao Yai Vineyard Stay','ปากช่อง นครราชสีมา',4.8,410,4800,['วิวไร่องุ่น','อาหารเช้า','สระส่วนตัว'])]
F2 = [('เปลี่ยนวันเดินทางได้หรือไม่?','ได้ หากรายการนั้นอนุญาต ไปที่ การจองของฉัน > รายละเอียด แล้วติดต่อทีมงานเพื่อแก้ไขวันที่'),
('คูปองใช้ร่วมกันได้หรือไม่?','ใช้ได้ครั้งละ 1 คูปองต่อ 1 การจอง และต้องมียอดขั้นต่ำตามเงื่อนไขของคูปอง'),
('เงินคืนเข้ากระเป๋าเมื่อไร?','เงินคืนเข้า Wallet ทันทีเมื่อการยกเลิกได้รับการยืนยัน'),
('ลืมรหัสผ่านต้องทำอย่างไร?','ไปที่ โปรไฟล์ > ความปลอดภัย > เปลี่ยนรหัสผ่าน หรือติดต่อทีมงานผ่านหน้า ติดต่อเรา'),
('ใบกำกับภาษีขอได้ที่ไหน?','ส่งคำขอผ่านหน้า ติดต่อเรา ระบุเลขที่การจอง ทีมงานจะดำเนินการภายใน 3 วันทำการ'),
('ข้อมูลส่วนตัวถูกใช้อย่างไร?','เราใช้ข้อมูลเพื่อให้บริการจองและปรับปรุงประสบการณ์เท่านั้น อ่านรายละเอียดได้ที่นโยบายความเป็นส่วนตัว'),
('ลบบัญชีได้หรือไม่?','ได้ที่ ตั้งค่า > ลบบัญชี ข้อมูลส่วนบุคคลจะถูกทำให้ไม่ระบุตัวตน'),
('จองให้ผู้อื่นได้หรือไม่?','ได้ ระบุจำนวนผู้เดินทางตอนจอง และแสดงเลขที่การจองเมื่อเช็กอิน'),
('PromptPay ชำระภายในกี่นาที?','QR มีอายุ 15 นาที หากหมดอายุสามารถสร้างรายการชำระใหม่ได้'),
('Wallet มีวันหมดอายุหรือไม่?','ยอดเงินใน Wallet ไม่มีวันหมดอายุ'),
('สมัครเป็นพาร์ทเนอร์อย่างไร?','ไปที่ เมนู > สมัครเป็นพาร์ทเนอร์ กรอกข้อมูลธุรกิจ ทีมงานจะพิจารณาภายใน 2 วันทำการ'),
('ค่าคอมมิชชันพาร์ทเนอร์เท่าไร?','ค่าเริ่มต้น 12% ของยอดรายการ ปรับได้ตามที่แอดมินกำหนด'),
('รีวิวเขียนได้เมื่อใด?','เขียนรีวิวได้หลังการจองเสร็จสิ้น 1 ครั้งต่อรายการ'),
('ติดต่อฉุกเฉินระหว่างเดินทาง?','ใช้หน้า ติดต่อเรา เลือกหัวข้อเร่งด่วน หรือโทรสายด่วน 1155 สำหรับตำรวจท่องเที่ยว'),
('AI วางแผนทริปแม่นยำแค่ไหน?','เป็นข้อเสนอแนะเบื้องต้น ควรตรวจสอบเวลาเปิดปิดและราคาก่อนเดินทาง'),
('ใช้แอปได้กี่ภาษา?','รองรับภาษาไทยและอังกฤษ ตั้งค่าได้ที่ ตั้งค่า > ภาษา')]
D, L, F = D0 + D2, L0 + L2, F0 + F2

_KIND = {'hotel': 'ที่พักคุณภาพ', 'restaurant': 'ร้านอาหารแนะนำ', 'attraction': 'กิจกรรมยอดนิยม', 'flight': 'เที่ยวบินตรง', 'trip': 'แพ็กเกจเที่ยวครบ จบในเที่ยวเดียว'}
def desc(i, t, n, l, am):
    if i <= 12:
        return 'รายละเอียดตัวอย่างของ ' + n
    return f"{n} {_KIND[t]}ที่{l} จุดเด่น: {' · '.join(am)} ยกเลิกฟรีตามเงื่อนไขของรายการ ยืนยันทันทีผ่านแอป"
q = lambda s: "'" + s.replace("'", "''") + "'"
gs = lambda s: json.dumps(s, ensure_ascii=False)
ds = lambda s: "'" + s.replace("\\", "\\\\").replace("'", "\\'").replace("$", "\\$") + "'"

def sql(D=D0, L=L0, F=F0, title="", faq_start=0):
    s = "-- DEV/BETA seed (synthetic data only). GENERATED by scripts/gen_seed.py — do not edit by hand.\n" + title
    s += "-- Applied only when MIGRATE_SEED=true and APP_ENV != production.\n"
    s += "INSERT INTO destinations (id,slug,name_th,subtitle,gradient_css) VALUES\n" + ",\n".join(f"({a},{q(b)},{q(c)},{q(d)},{q(e)})" for a,b,c,d,e in D) + ";\n"
    s += "INSERT INTO listings (id,destination_id,type,name,location_text,description,rating_avg,review_count,price_from,amenities,map_query,hero_image_key) VALUES\n" + ",\n".join(
        f"({i},{d},{q(t)},{q(n)},{q(l)},{q(desc(i,t,n,l,am))},{r},{c},{p},{q(json.dumps(am,ensure_ascii=False))},{q(n+' '+l)},{q('seed/listing-%d.jpg'%i)})"
        for i,d,t,n,l,r,c,p,am in L) + ";\n"
    s += "INSERT INTO faqs (question,answer,sort_no) VALUES\n" + ",\n".join(f"({q(a)},{q(b)},{i})" for i,(a,b) in enumerate(F, faq_start)) + ";\n"
    if not faq_start:
        s += "INSERT INTO coupons (code,title,detail,percent_off,min_spend,max_discount,expires_on) VALUES\n('WELCOME10','ส่วนลด 10%','สำหรับการจองครั้งแรก',10,1000.00,500.00,'2027-12-31');\n"
    return s

def go():
    g = "// Code generated by scripts/gen_seed.py; DO NOT EDIT.\npackage api\n\n"
    g += "var seedDestinations = []Destination{\n" + "".join(f"\t{{ID: {a}, Slug: {gs(b)}, Name: {gs(c)}, Subtitle: {gs(d)}, GradientCSS: {gs(e)}}},\n" for a,b,c,d,e in D) + "}\n\n"
    g += "var seedListings = []listingRow{\n" + "".join(
        f"\t{{ListingDetail: ListingDetail{{ListingSummary: ListingSummary{{ID: {i}, Type: {gs(t)}, Name: {gs(n)}, Location: {gs(l)}, Rating: {r}, ReviewCount: {c}, PriceFrom: {p}}}, Description: {gs(desc(i,t,n,l,am))}, Amenities: []string{{{', '.join(gs(x) for x in am)}}}, MapQuery: {gs(n+' '+l)}}}, DestinationID: {d}}},\n"
        for i,d,t,n,l,r,c,p,am in L) + "}\n\n"
    g += "var seedFaqs = []Faq{\n" + "".join(f"\t{{ID: {i+1}, Question: {gs(a)}, Answer: {gs(b)}}},\n" for i,(a,b) in enumerate(F)) + "}\n"
    return g

def dart():
    s = "// GENERATED by scripts/gen_seed.py; DO NOT EDIT. Mirrors services/api-go seed data.\n"
    s += "const List<Map<String, Object?>> kMockDestinations = [\n" + "".join(
        f"  {{'id': {a}, 'slug': {ds(b)}, 'name': {ds(c)}, 'subtitle': {ds(d)}, 'gradientCss': {ds(e)}}},\n" for a,b,c,d,e in D) + "];\n\n"
    s += "const List<Map<String, Object?>> kMockListings = [\n" + "".join(
        f"  {{'id': {i}, 'destinationId': {d}, 'type': {ds(t)}, 'name': {ds(n)}, 'location': {ds(l)}, 'rating': {r}, 'reviewCount': {c}, 'priceFrom': {float(p)}, "
        f"'description': {ds(desc(i,t,n,l,am))}, 'amenities': [{', '.join(ds(x) for x in am)}], 'mapQuery': {ds(n+' '+l)}}},\n"
        for i,d,t,n,l,r,c,p,am in L) + "];\n\n"
    s += "const List<Map<String, Object?>> kMockFaqs = [\n" + "".join(f"  {{'id': {i+1}, 'question': {ds(a)}, 'answer': {ds(b)}}},\n" for i,(a,b) in enumerate(F)) + "];\n"
    return s

OUT = {
    "services/api-go/migrations/0002_seed_dev.up.sql": sql(),
    "services/api-go/migrations/0006_seed_demo_catalog.up.sql": sql(D2, L2, F2, "-- Demo catalog extension: 2 destinations, 18 listings, 16 FAQs.\n", len(F0)),
    "services/api-go/internal/api/fixtures.go": go(),
    "apps/mobile-flutter/lib/data/mock_fixtures.dart": dart(),
}
if __name__ == "__main__":
    check = "--check" in sys.argv
    bad = 0
    for rel, content in OUT.items():
        p = ROOT / rel
        if check:
            norm = lambda t: "".join(t.split())  # gofmt may re-align whitespace
            if not p.exists() or norm(p.read_text()) != norm(content):
                print("DRIFT:", rel); bad += 1
        else:
            p.parent.mkdir(parents=True, exist_ok=True)
            p.write_text(content)
            print("wrote", rel)
    sys.exit(1 if bad else 0)
