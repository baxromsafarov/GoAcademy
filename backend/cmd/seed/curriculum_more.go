package main

// init appends a second wave of topics to the base curriculum so each language
// ships ~20 articles/quizzes/problems instead of 8. Kept in its own file (and
// appended via init) so the base curriculum literal stays untouched.
func init() {
	curriculum = append(curriculum, moreTopics...)
}

var moreTopics = []topic{
	{
		Tag: "go-strings", Difficulty: "beginner",
		Title: ml{"ru": "Строки и руны", "en": "Strings and runes", "uz": "Satrlar va runalar", "ja": "文字列とルーン"},
		Blurb: ml{"ru": "Строки в UTF-8, байты и руны.", "en": "UTF-8 strings, bytes and runes.", "uz": "UTF-8 satrlar, baytlar va runalar.", "ja": "UTF-8 文字列、バイト、ルーン。"},
		Article: ml{
			"en": "# Strings and runes\n\nA Go `string` is read-only UTF-8 bytes. Range over it to get runes (code points).\n\n```go\nfor i, r := range \"go\" {\n\tfmt.Println(i, string(r))\n}\n```",
			"ru": "# Строки и руны\n\n`string` в Go — это неизменяемые байты в UTF-8. Перебор `range` даёт руны (кодовые точки).\n\n```go\nfor i, r := range \"go\" {\n\tfmt.Println(i, string(r))\n}\n```",
			"uz": "# Satrlar va runalar\n\nGo'dagi `string` — UTF-8 baytlar (o'zgarmas). `range` runalarni (kod nuqtalarini) beradi.\n\n```go\nfor i, r := range \"go\" {\n\tfmt.Println(i, string(r))\n}\n```",
			"ja": "# 文字列とルーン\n\nGo の `string` は読み取り専用の UTF-8 バイト列です。`range` でルーン（コードポイント）を得られます。\n\n```go\nfor i, r := range \"go\" {\n\tfmt.Println(i, string(r))\n}\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что даёт `range` по строке?", "en": "What does ranging over a string yield?", "uz": "Satr bo'yicha `range` nimani beradi?", "ja": "文字列を range すると何が得られる？"},
				opt(ml{"ru": "Руны (код. точки)", "en": "Runes (code points)", "uz": "Runalar (kod nuqtalari)", "ja": "ルーン（コードポイント）"}, true),
				opt(ml{"ru": "Всегда байты", "en": "Always bytes", "uz": "Doim baytlar", "ja": "常にバイト"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-reverse", Title: ml{"ru": "Переворот строки", "en": "Reverse a string", "uz": "Satrni teskari", "ja": "文字列の反転"},
			Statement: ml{"ru": "Прочитайте строку и выведите её в обратном порядке.", "en": "Read a line and print it reversed.", "uz": "Bir qatorni o'qing va teskari chiqaring.", "ja": "1 行を読み取り、逆順で出力してください。"},
			Solution:  "```go\npackage main\nimport (\"bufio\";\"fmt\";\"os\";\"strings\")\nfunc main(){ s,_:=bufio.NewReader(os.Stdin).ReadString('\\n'); s=strings.TrimRight(s,\"\\n\"); r:=[]rune(s); for i,j:=0,len(r)-1;i<j;i,j=i+1,j-1{ r[i],r[j]=r[j],r[i] }; fmt.Println(string(r)) }\n```",
			Cases:     []testCase{{In: "abc\n", Out: "cba", Sample: true}, {In: "go\n", Out: "og", Sample: false}},
		},
	},
	{
		Tag: "go-pointers", Difficulty: "beginner",
		Title: ml{"ru": "Указатели", "en": "Pointers", "uz": "Ko'rsatkichlar", "ja": "ポインタ"},
		Blurb: ml{"ru": "Адреса значений: & и *.", "en": "Addresses of values: & and *.", "uz": "Qiymat manzillari: & va *.", "ja": "値のアドレス: & と *。"},
		Article: ml{
			"en": "# Pointers\n\n`&x` takes the address of `x`; `*p` dereferences a pointer. Passing a pointer lets a function modify the caller's value.\n\n```go\nfunc inc(p *int) { *p++ }\nn := 1\ninc(&n) // n == 2\n```",
			"ru": "# Указатели\n\n`&x` берёт адрес `x`; `*p` разыменовывает указатель. Передача указателя позволяет функции изменить значение вызывающего.\n\n```go\nfunc inc(p *int) { *p++ }\nn := 1\ninc(&n) // n == 2\n```",
			"uz": "# Ko'rsatkichlar\n\n`&x` — `x` manzili; `*p` — ko'rsatkichni ochish. Ko'rsatkich uzatish funksiyaga chaqiruvchi qiymatini o'zgartirishga imkon beradi.\n\n```go\nfunc inc(p *int) { *p++ }\nn := 1\ninc(&n) // n == 2\n```",
			"ja": "# ポインタ\n\n`&x` は `x` のアドレス、`*p` はデリファレンスです。ポインタを渡すと関数が呼び出し元の値を変更できます。\n\n```go\nfunc inc(p *int) { *p++ }\nn := 1\ninc(&n) // n == 2\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что делает `*p`?", "en": "What does `*p` do?", "uz": "`*p` nima qiladi?", "ja": "`*p` は何をする？"},
				opt(ml{"ru": "Разыменовывает указатель", "en": "Dereferences the pointer", "uz": "Ko'rsatkichni ochadi", "ja": "ポインタをデリファレンスする"}, true),
				opt(ml{"ru": "Берёт адрес", "en": "Takes an address", "uz": "Manzilni oladi", "ja": "アドレスを取る"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-double", Title: ml{"ru": "Удвоение", "en": "Double", "uz": "Ikkilantirish", "ja": "2 倍"},
			Statement: ml{"ru": "Прочитайте целое число и выведите его удвоенное значение.", "en": "Read an integer and print double its value.", "uz": "Butun sonni o'qing va ikki barobarini chiqaring.", "ja": "整数を読み取り、その 2 倍を出力してください。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var n int; fmt.Scan(&n); fmt.Println(n*2) }\n```",
			Cases:     []testCase{{In: "4\n", Out: "8", Sample: true}, {In: "-3\n", Out: "-6", Sample: false}},
		},
	},
	{
		Tag: "go-errors", Difficulty: "beginner",
		Title: ml{"ru": "Ошибки", "en": "Errors", "uz": "Xatolar", "ja": "エラー"},
		Blurb: ml{"ru": "Идиома error как значение.", "en": "The error-as-value idiom.", "uz": "Xato-qiymat idiomasi.", "ja": "値としてのエラー。"},
		Article: ml{
			"en": "# Errors\n\nErrors are ordinary values of type `error`. Check them with `if err != nil`.\n\n```go\nv, err := strconv.Atoi(\"x\")\nif err != nil {\n\tfmt.Println(\"bad number\")\n}\n_ = v\n```",
			"ru": "# Ошибки\n\nОшибки — обычные значения типа `error`. Проверяйте `if err != nil`.\n\n```go\nv, err := strconv.Atoi(\"x\")\nif err != nil {\n\tfmt.Println(\"bad number\")\n}\n_ = v\n```",
			"uz": "# Xatolar\n\nXatolar — `error` turidagi oddiy qiymatlar. `if err != nil` bilan tekshiring.\n\n```go\nv, err := strconv.Atoi(\"x\")\nif err != nil {\n\tfmt.Println(\"bad number\")\n}\n_ = v\n```",
			"ja": "# エラー\n\nエラーは `error` 型の普通の値です。`if err != nil` で確認します。\n\n```go\nv, err := strconv.Atoi(\"x\")\nif err != nil {\n\tfmt.Println(\"bad number\")\n}\n_ = v\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Как проверяют ошибку в Go?", "en": "How do you check an error in Go?", "uz": "Go'da xato qanday tekshiriladi?", "ja": "Go でエラーをどう確認する？"},
				opt(ml{"ru": "if err != nil", "en": "if err != nil", "uz": "if err != nil", "ja": "if err != nil"}, true),
				opt(ml{"ru": "try/catch", "en": "try/catch", "uz": "try/catch", "ja": "try/catch"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-safediv", Title: ml{"ru": "Безопасное деление", "en": "Safe division", "uz": "Xavfsiz bo'lish", "ja": "安全な除算"},
			Statement: ml{"ru": "Прочитайте a и b. Если b равно 0, выведите error, иначе a/b (целочисленно).", "en": "Read a and b. If b is 0 print error, otherwise a/b (integer).", "uz": "a va b ni o'qing. Agar b 0 bo'lsa error, aks holda a/b (butun).", "ja": "a と b を読み取り、b が 0 なら error、そうでなければ a/b（整数）を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var a,b int; fmt.Scan(&a,&b); if b==0{ fmt.Println(\"error\") } else { fmt.Println(a/b) } }\n```",
			Cases:     []testCase{{In: "6 3\n", Out: "2", Sample: true}, {In: "5 0\n", Out: "error", Sample: false}},
		},
	},
	{
		Tag: "go-arrays", Difficulty: "beginner",
		Title: ml{"ru": "Массивы", "en": "Arrays", "uz": "Massivlar", "ja": "配列"},
		Blurb: ml{"ru": "Массивы фиксированной длины.", "en": "Fixed-length arrays.", "uz": "Belgilangan uzunlikdagi massivlar.", "ja": "固定長の配列。"},
		Article: ml{
			"en": "# Arrays\n\nAn array `[n]T` has a fixed length that is part of its type. Most Go code prefers slices, but arrays are useful for fixed data.\n\n```go\nvar a [3]int\na[0] = 10\nfmt.Println(len(a)) // 3\n```",
			"ru": "# Массивы\n\nМассив `[n]T` имеет фиксированную длину, которая входит в тип. Чаще используют срезы, но массивы удобны для фиксированных данных.\n\n```go\nvar a [3]int\na[0] = 10\nfmt.Println(len(a)) // 3\n```",
			"uz": "# Massivlar\n\n`[n]T` massivining uzunligi belgilangan va turning bir qismidir. Ko'pincha slice afzal, lekin massivlar belgilangan ma'lumot uchun qulay.\n\n```go\nvar a [3]int\na[0] = 10\nfmt.Println(len(a)) // 3\n```",
			"ja": "# 配列\n\n配列 `[n]T` は長さが型の一部で固定です。多くの場合スライスが好まれますが、固定データには配列が便利です。\n\n```go\nvar a [3]int\na[0] = 10\nfmt.Println(len(a)) // 3\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Длина массива в Go...", "en": "An array's length in Go is...", "uz": "Go'da massiv uzunligi...", "ja": "Go の配列の長さは…"},
				opt(ml{"ru": "Часть его типа", "en": "Part of its type", "uz": "Turning bir qismi", "ja": "型の一部"}, true),
				opt(ml{"ru": "Меняется в рантайме", "en": "Changeable at runtime", "uz": "Runtime'da o'zgaradi", "ja": "実行時に変えられる"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-arraysum", Title: ml{"ru": "Сумма массива", "en": "Array sum", "uz": "Massiv yig'indisi", "ja": "配列の合計"},
			Statement: ml{"ru": "Прочитайте n и n чисел, выведите их сумму.", "en": "Read n then n integers, print their sum.", "uz": "n va n ta sonni o'qing, yig'indisini chiqaring.", "ja": "n と n 個の整数を読み取り、合計を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var n int; fmt.Scan(&n); sum:=0; for i:=0;i<n;i++{ var x int; fmt.Scan(&x); sum+=x }; fmt.Println(sum) }\n```",
			Cases:     []testCase{{In: "3\n1 2 3\n", Out: "6", Sample: true}, {In: "4\n10 20 30 40\n", Out: "100", Sample: false}},
		},
	},
	{
		Tag: "go-packages", Difficulty: "beginner",
		Title: ml{"ru": "Пакеты и импорт", "en": "Packages and imports", "uz": "Paketlar va import", "ja": "パッケージとインポート"},
		Blurb: ml{"ru": "Организация кода в пакеты.", "en": "Organising code into packages.", "uz": "Kodni paketlarga ajratish.", "ja": "コードをパッケージに整理する。"},
		Article: ml{
			"en": "# Packages and imports\n\nEvery file declares a `package`. Import others by path; the standard library is rich — e.g. `strings`.\n\n```go\nimport \"strings\"\nfmt.Println(strings.ToUpper(\"go\")) // GO\n```",
			"ru": "# Пакеты и импорт\n\nКаждый файл объявляет `package`. Другие подключают через путь; стандартная библиотека богата — например `strings`.\n\n```go\nimport \"strings\"\nfmt.Println(strings.ToUpper(\"go\")) // GO\n```",
			"uz": "# Paketlar va import\n\nHar bir fayl `package` e'lon qiladi. Boshqalarini yo'l orqali import qiling; standart kutubxona boy — masalan `strings`.\n\n```go\nimport \"strings\"\nfmt.Println(strings.ToUpper(\"go\")) // GO\n```",
			"ja": "# パッケージとインポート\n\nすべてのファイルは `package` を宣言します。他はパスでインポートします。標準ライブラリは豊富で、例えば `strings`。\n\n```go\nimport \"strings\"\nfmt.Println(strings.ToUpper(\"go\")) // GO\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Откуда импортируют strings.ToUpper?", "en": "Where does strings.ToUpper come from?", "uz": "strings.ToUpper qayerdan?", "ja": "strings.ToUpper はどこから？"},
				opt(ml{"ru": "Стандартная библиотека", "en": "The standard library", "uz": "Standart kutubxona", "ja": "標準ライブラリ"}, true),
				opt(ml{"ru": "Сторонний пакет", "en": "A third-party package", "uz": "Uchinchi tomon paketi", "ja": "サードパーティ"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-upper", Title: ml{"ru": "В верхний регистр", "en": "To upper case", "uz": "Katta harflarga", "ja": "大文字へ"},
			Statement: ml{"ru": "Прочитайте строку и выведите её в верхнем регистре.", "en": "Read a line and print it upper-cased.", "uz": "Bir qatorni o'qing va katta harflarda chiqaring.", "ja": "1 行を読み取り、大文字で出力してください。"},
			Solution:  "```go\npackage main\nimport (\"bufio\";\"fmt\";\"os\";\"strings\")\nfunc main(){ s,_:=bufio.NewReader(os.Stdin).ReadString('\\n'); fmt.Println(strings.ToUpper(strings.TrimRight(s,\"\\n\"))) }\n```",
			Cases:     []testCase{{In: "go\n", Out: "GO", Sample: true}, {In: "Hello\n", Out: "HELLO", Sample: false}},
		},
	},
	{
		Tag: "go-closures", Difficulty: "intermediate",
		Title: ml{"ru": "Замыкания", "en": "Closures", "uz": "Closure'lar", "ja": "クロージャ"},
		Blurb: ml{"ru": "Функции, захватывающие переменные.", "en": "Functions that capture variables.", "uz": "O'zgaruvchini ushlaydigan funksiyalar.", "ja": "変数を捕捉する関数。"},
		Article: ml{
			"en": "# Closures\n\nA function literal can capture variables from its scope, keeping state between calls.\n\n```go\nfunc counter() func() int {\n\tn := 0\n\treturn func() int { n++; return n }\n}\n```",
			"ru": "# Замыкания\n\nФункция-литерал может захватывать переменные области видимости, сохраняя состояние между вызовами.\n\n```go\nfunc counter() func() int {\n\tn := 0\n\treturn func() int { n++; return n }\n}\n```",
			"uz": "# Closure'lar\n\nFunksiya-literal o'z doirasidagi o'zgaruvchilarni ushlab, chaqiruvlar orasida holatni saqlaydi.\n\n```go\nfunc counter() func() int {\n\tn := 0\n\treturn func() int { n++; return n }\n}\n```",
			"ja": "# クロージャ\n\n関数リテラルはスコープの変数を捕捉し、呼び出し間で状態を保持できます。\n\n```go\nfunc counter() func() int {\n\tn := 0\n\treturn func() int { n++; return n }\n}\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что делает замыкание?", "en": "What does a closure do?", "uz": "Closure nima qiladi?", "ja": "クロージャは何をする？"},
				opt(ml{"ru": "Захватывает переменные окружения", "en": "Captures surrounding variables", "uz": "Atrofdagi o'zgaruvchilarni ushlaydi", "ja": "周囲の変数を捕捉する"}, true),
				opt(ml{"ru": "Запрещает рекурсию", "en": "Forbids recursion", "uz": "Rekursiyani taqiqlaydi", "ja": "再帰を禁止する"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-range1n", Title: ml{"ru": "Ряд 1..n", "en": "Sequence 1..n", "uz": "1..n ketma-ketlik", "ja": "数列 1..n"},
			Statement: ml{"ru": "Прочитайте n и выведите числа 1..n через пробел в одной строке.", "en": "Read n and print 1..n space-separated on one line.", "uz": "n ni o'qing va 1..n sonlarni bir qatorda bo'sh joy bilan chiqaring.", "ja": "n を読み取り、1..n を空白区切りで 1 行に出力。"},
			Solution:  "```go\npackage main\nimport (\"fmt\";\"strings\";\"strconv\")\nfunc main(){ var n int; fmt.Scan(&n); gen:=func() func() string{ i:=0; return func() string{ i++; return strconv.Itoa(i) } }(); p:=make([]string,0,n); for k:=0;k<n;k++{ p=append(p,gen()) }; fmt.Println(strings.Join(p,\" \")) }\n```",
			Cases:     []testCase{{In: "3\n", Out: "1 2 3", Sample: true}, {In: "5\n", Out: "1 2 3 4 5", Sample: false}},
		},
	},
	{
		Tag: "go-defer", Difficulty: "intermediate",
		Title: ml{"ru": "defer, panic, recover", "en": "defer, panic, recover", "uz": "defer, panic, recover", "ja": "defer, panic, recover"},
		Blurb: ml{"ru": "Отложенный вызов и восстановление.", "en": "Deferred calls and recovery.", "uz": "Kechiktirilgan chaqiruv va tiklash.", "ja": "遅延呼び出しと回復。"},
		Article: ml{
			"en": "# defer, panic, recover\n\n`defer` schedules a call for when the function returns (LIFO order) — great for cleanup.\n\n```go\nf, _ := os.Open(name)\ndefer f.Close()\n```",
			"ru": "# defer, panic, recover\n\n`defer` откладывает вызов до выхода из функции (порядок LIFO) — удобно для очистки.\n\n```go\nf, _ := os.Open(name)\ndefer f.Close()\n```",
			"uz": "# defer, panic, recover\n\n`defer` chaqiruvni funksiya tugaganda bajaradi (LIFO tartibi) — tozalash uchun qulay.\n\n```go\nf, _ := os.Open(name)\ndefer f.Close()\n```",
			"ja": "# defer, panic, recover\n\n`defer` は関数の終了時に呼び出しを予約します（LIFO 順）。後始末に最適です。\n\n```go\nf, _ := os.Open(name)\ndefer f.Close()\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Когда выполняется defer?", "en": "When does a deferred call run?", "uz": "defer qachon bajariladi?", "ja": "defer はいつ実行される？"},
				opt(ml{"ru": "При выходе из функции", "en": "When the function returns", "uz": "Funksiya tugaganda", "ja": "関数が戻るとき"}, true),
				opt(ml{"ru": "Немедленно", "en": "Immediately", "uz": "Darhol", "ja": "すぐに"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-factorial", Title: ml{"ru": "Факториал", "en": "Factorial", "uz": "Faktorial", "ja": "階乗"},
			Statement: ml{"ru": "Прочитайте n (0..12) и выведите n!.", "en": "Read n (0..12) and print n!.", "uz": "n (0..12) ni o'qing va n! ni chiqaring.", "ja": "n (0..12) を読み取り n! を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var n int; fmt.Scan(&n); f:=1; for i:=2;i<=n;i++{ f*=i }; fmt.Println(f) }\n```",
			Cases:     []testCase{{In: "5\n", Out: "120", Sample: true}, {In: "0\n", Out: "1", Sample: false}},
		},
	},
	{
		Tag: "go-json", Difficulty: "intermediate",
		Title: ml{"ru": "JSON", "en": "JSON", "uz": "JSON", "ja": "JSON"},
		Blurb: ml{"ru": "Кодирование и декодирование JSON.", "en": "Encoding and decoding JSON.", "uz": "JSON kodlash va dekodlash.", "ja": "JSON のエンコードとデコード。"},
		Article: ml{
			"en": "# JSON\n\n`encoding/json` marshals structs to JSON. Field tags control the keys.\n\n```go\ntype P struct{ Sum int `json:\"sum\"` }\nb, _ := json.Marshal(P{5})\nfmt.Println(string(b)) // {\"sum\":5}\n```",
			"ru": "# JSON\n\n`encoding/json` сериализует структуры в JSON. Теги полей задают ключи.\n\n```go\ntype P struct{ Sum int `json:\"sum\"` }\nb, _ := json.Marshal(P{5})\nfmt.Println(string(b)) // {\"sum\":5}\n```",
			"uz": "# JSON\n\n`encoding/json` structlarni JSON'ga aylantiradi. Maydon teglari kalitlarni belgilaydi.\n\n```go\ntype P struct{ Sum int `json:\"sum\"` }\nb, _ := json.Marshal(P{5})\nfmt.Println(string(b)) // {\"sum\":5}\n```",
			"ja": "# JSON\n\n`encoding/json` は構造体を JSON に変換します。フィールドタグでキーを制御します。\n\n```go\ntype P struct{ Sum int `json:\"sum\"` }\nb, _ := json.Marshal(P{5})\nfmt.Println(string(b)) // {\"sum\":5}\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что задаёт имя JSON-ключа?", "en": "What sets the JSON key name?", "uz": "JSON kalit nomini nima belgilaydi?", "ja": "JSON キー名を決めるのは？"},
				opt(ml{"ru": "Тег поля структуры", "en": "A struct field tag", "uz": "Struct maydon tegi", "ja": "構造体フィールドタグ"}, true),
				opt(ml{"ru": "Имя файла", "en": "The file name", "uz": "Fayl nomi", "ja": "ファイル名"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-jsonsum", Title: ml{"ru": "JSON-сумма", "en": "JSON sum", "uz": "JSON yig'indi", "ja": "JSON 合計"},
			Statement: ml{"ru": "Прочитайте a и b, выведите {\"sum\":a+b} как JSON.", "en": "Read a and b, print {\"sum\":a+b} as JSON.", "uz": "a va b ni o'qing, {\"sum\":a+b} ni JSON sifatida chiqaring.", "ja": "a と b を読み取り、{\"sum\":a+b} を JSON で出力。"},
			Solution:  "```go\npackage main\nimport (\"encoding/json\";\"fmt\")\nfunc main(){ var a,b int; fmt.Scan(&a,&b); out:=struct{ Sum int `json:\"sum\"` }{a+b}; bs,_:=json.Marshal(out); fmt.Println(string(bs)) }\n```",
			Cases:     []testCase{{In: "2 3\n", Out: "{\"sum\":5}", Sample: true}, {In: "10 5\n", Out: "{\"sum\":15}", Sample: false}},
		},
	},
	{
		Tag: "go-sorting", Difficulty: "intermediate",
		Title: ml{"ru": "Сортировка", "en": "Sorting", "uz": "Saralash", "ja": "ソート"},
		Blurb: ml{"ru": "Пакет sort.", "en": "The sort package.", "uz": "sort paketi.", "ja": "sort パッケージ。"},
		Article: ml{
			"en": "# Sorting\n\nThe `sort` package sorts slices in place.\n\n```go\ns := []int{3, 1, 2}\nsort.Ints(s) // [1 2 3]\n```",
			"ru": "# Сортировка\n\nПакет `sort` сортирует срезы на месте.\n\n```go\ns := []int{3, 1, 2}\nsort.Ints(s) // [1 2 3]\n```",
			"uz": "# Saralash\n\n`sort` paketi slice'larni joyida saralaydi.\n\n```go\ns := []int{3, 1, 2}\nsort.Ints(s) // [1 2 3]\n```",
			"ja": "# ソート\n\n`sort` パッケージはスライスをその場でソートします。\n\n```go\ns := []int{3, 1, 2}\nsort.Ints(s) // [1 2 3]\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что сортирует sort.Ints?", "en": "What does sort.Ints sort?", "uz": "sort.Ints nimani saralaydi?", "ja": "sort.Ints は何をソートする？"},
				opt(ml{"ru": "Срез []int на месте", "en": "An []int slice in place", "uz": "[]int slice'ni joyida", "ja": "[]int スライスをその場で"}, true),
				opt(ml{"ru": "Только строки", "en": "Only strings", "uz": "Faqat satrlar", "ja": "文字列のみ"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-sortn", Title: ml{"ru": "Сортировка чисел", "en": "Sort numbers", "uz": "Sonlarni saralash", "ja": "数値のソート"},
			Statement: ml{"ru": "Прочитайте n и n чисел, выведите их по возрастанию через пробел.", "en": "Read n then n integers, print them ascending, space-separated.", "uz": "n va n ta sonni o'qing, o'sish tartibida bo'sh joy bilan chiqaring.", "ja": "n と n 個の整数を読み取り、昇順で空白区切り出力。"},
			Solution:  "```go\npackage main\nimport (\"fmt\";\"sort\";\"strconv\";\"strings\")\nfunc main(){ var n int; fmt.Scan(&n); a:=make([]int,n); for i:=range a{ fmt.Scan(&a[i]) }; sort.Ints(a); p:=make([]string,n); for i,v:=range a{ p[i]=strconv.Itoa(v) }; fmt.Println(strings.Join(p,\" \")) }\n```",
			Cases:     []testCase{{In: "3\n3 1 2\n", Out: "1 2 3", Sample: true}, {In: "4\n9 7 8 6\n", Out: "6 7 8 9", Sample: false}},
		},
	},
	{
		Tag: "go-generics", Difficulty: "advanced",
		Title: ml{"ru": "Дженерики", "en": "Generics", "uz": "Generiklar", "ja": "ジェネリクス"},
		Blurb: ml{"ru": "Параметры типов.", "en": "Type parameters.", "uz": "Tur parametrlari.", "ja": "型パラメータ。"},
		Article: ml{
			"en": "# Generics\n\nType parameters let one function work over many types.\n\n```go\nfunc Max[T int | float64](a, b T) T {\n\tif a > b { return a }\n\treturn b\n}\n```",
			"ru": "# Дженерики\n\nПараметры типов позволяют одной функции работать со многими типами.\n\n```go\nfunc Max[T int | float64](a, b T) T {\n\tif a > b { return a }\n\treturn b\n}\n```",
			"uz": "# Generiklar\n\nTur parametrlari bitta funksiyani ko'p turlar bilan ishlatishga imkon beradi.\n\n```go\nfunc Max[T int | float64](a, b T) T {\n\tif a > b { return a }\n\treturn b\n}\n```",
			"ja": "# ジェネリクス\n\n型パラメータで 1 つの関数が多くの型に対応できます。\n\n```go\nfunc Max[T int | float64](a, b T) T {\n\tif a > b { return a }\n\treturn b\n}\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что добавляют дженерики?", "en": "What do generics add?", "uz": "Generiklar nima qo'shadi?", "ja": "ジェネリクスは何を加える？"},
				opt(ml{"ru": "Параметры типов", "en": "Type parameters", "uz": "Tur parametrlari", "ja": "型パラメータ"}, true),
				opt(ml{"ru": "Динамическую типизацию", "en": "Dynamic typing", "uz": "Dinamik turlash", "ja": "動的型付け"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-genmax", Title: ml{"ru": "Максимум (дженерик)", "en": "Maximum (generic)", "uz": "Maksimum (generik)", "ja": "最大値（ジェネリック）"},
			Statement: ml{"ru": "Прочитайте n и n чисел, выведите максимум.", "en": "Read n then n integers, print the maximum.", "uz": "n va n ta sonni o'qing, maksimumni chiqaring.", "ja": "n と n 個の整数を読み取り、最大値を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc Max[T int](a,b T) T { if a>b { return a }; return b }\nfunc main(){ var n int; fmt.Scan(&n); var m int; for i:=0;i<n;i++{ var x int; fmt.Scan(&x); if i==0||Max(x,m)==x { m=x } }; fmt.Println(m) }\n```",
			Cases:     []testCase{{In: "3\n1 5 2\n", Out: "5", Sample: true}, {In: "4\n-3 -1 -9 -2\n", Out: "-1", Sample: false}},
		},
	},
	{
		Tag: "go-sync", Difficulty: "advanced",
		Title: ml{"ru": "sync: WaitGroup и Mutex", "en": "sync: WaitGroup and Mutex", "uz": "sync: WaitGroup va Mutex", "ja": "sync: WaitGroup と Mutex"},
		Blurb: ml{"ru": "Ожидание горутин и защита данных.", "en": "Waiting on goroutines and guarding data.", "uz": "Goroutine'larni kutish va ma'lumotni himoyalash.", "ja": "ゴルーチンの待機とデータ保護。"},
		Article: ml{
			"en": "# sync: WaitGroup and Mutex\n\n`sync.WaitGroup` waits for goroutines; `sync.Mutex` guards shared state.\n\n```go\nvar wg sync.WaitGroup\nwg.Add(1)\ngo func() { defer wg.Done() }()\nwg.Wait()\n```",
			"ru": "# sync: WaitGroup и Mutex\n\n`sync.WaitGroup` ждёт горутины; `sync.Mutex` защищает общие данные.\n\n```go\nvar wg sync.WaitGroup\nwg.Add(1)\ngo func() { defer wg.Done() }()\nwg.Wait()\n```",
			"uz": "# sync: WaitGroup va Mutex\n\n`sync.WaitGroup` goroutine'larni kutadi; `sync.Mutex` umumiy ma'lumotni himoyalaydi.\n\n```go\nvar wg sync.WaitGroup\nwg.Add(1)\ngo func() { defer wg.Done() }()\nwg.Wait()\n```",
			"ja": "# sync: WaitGroup と Mutex\n\n`sync.WaitGroup` はゴルーチンを待ち、`sync.Mutex` は共有状態を守ります。\n\n```go\nvar wg sync.WaitGroup\nwg.Add(1)\ngo func() { defer wg.Done() }()\nwg.Wait()\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Зачем нужен sync.WaitGroup?", "en": "What is sync.WaitGroup for?", "uz": "sync.WaitGroup nima uchun?", "ja": "sync.WaitGroup は何のため？"},
				opt(ml{"ru": "Ждать завершения горутин", "en": "To wait for goroutines to finish", "uz": "Goroutine tugashini kutish", "ja": "ゴルーチンの完了を待つ"}, true),
				opt(ml{"ru": "Ускорять цикл for", "en": "To speed up a for loop", "uz": "for siklni tezlashtirish", "ja": "for ループを高速化する"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-concsum", Title: ml{"ru": "Сумма (горутины)", "en": "Sum (goroutines)", "uz": "Yig'indi (goroutine)", "ja": "合計（ゴルーチン）"},
			Statement: ml{"ru": "Прочитайте n и n чисел, выведите их сумму.", "en": "Read n then n integers, print their sum.", "uz": "n va n ta sonni o'qing, yig'indisini chiqaring.", "ja": "n と n 個の整数を読み取り、合計を出力。"},
			Solution:  "```go\npackage main\nimport (\"fmt\";\"sync\")\nfunc main(){ var n int; fmt.Scan(&n); a:=make([]int,n); for i:=range a{ fmt.Scan(&a[i]) }; var wg sync.WaitGroup; var mu sync.Mutex; sum:=0; for _,v:=range a{ wg.Add(1); go func(x int){ defer wg.Done(); mu.Lock(); sum+=x; mu.Unlock() }(v) }; wg.Wait(); fmt.Println(sum) }\n```",
			Cases:     []testCase{{In: "3\n1 2 3\n", Out: "6", Sample: true}, {In: "4\n5 5 5 5\n", Out: "20", Sample: false}},
		},
	},
	{
		Tag: "go-select", Difficulty: "advanced",
		Title: ml{"ru": "Оператор select", "en": "The select statement", "uz": "select operatori", "ja": "select 文"},
		Blurb: ml{"ru": "Ожидание нескольких каналов.", "en": "Waiting on multiple channels.", "uz": "Bir nechta kanalni kutish.", "ja": "複数チャネルの待機。"},
		Article: ml{
			"en": "# The select statement\n\n`select` waits on multiple channel operations and runs the first that is ready.\n\n```go\nselect {\ncase v := <-a:\n\tfmt.Println(v)\ncase v := <-b:\n\tfmt.Println(v)\n}\n```",
			"ru": "# Оператор select\n\n`select` ждёт несколько канальных операций и выполняет первую готовую.\n\n```go\nselect {\ncase v := <-a:\n\tfmt.Println(v)\ncase v := <-b:\n\tfmt.Println(v)\n}\n```",
			"uz": "# select operatori\n\n`select` bir nechta kanal amalini kutadi va birinchi tayyorini bajaradi.\n\n```go\nselect {\ncase v := <-a:\n\tfmt.Println(v)\ncase v := <-b:\n\tfmt.Println(v)\n}\n```",
			"ja": "# select 文\n\n`select` は複数のチャネル操作を待ち、最初に準備できたものを実行します。\n\n```go\nselect {\ncase v := <-a:\n\tfmt.Println(v)\ncase v := <-b:\n\tfmt.Println(v)\n}\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что делает select?", "en": "What does select do?", "uz": "select nima qiladi?", "ja": "select は何をする？"},
				opt(ml{"ru": "Ждёт несколько каналов", "en": "Waits on multiple channels", "uz": "Bir nechta kanalni kutadi", "ja": "複数チャネルを待つ"}, true),
				opt(ml{"ru": "Объявляет структуру", "en": "Declares a struct", "uz": "Struct e'lon qiladi", "ja": "構造体を宣言する"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-min2", Title: ml{"ru": "Минимум из двух", "en": "Minimum of two", "uz": "Ikkidan minimum", "ja": "2 つの最小値"},
			Statement: ml{"ru": "Прочитайте a и b, выведите меньшее.", "en": "Read a and b, print the smaller.", "uz": "a va b ni o'qing, kichigini chiqaring.", "ja": "a と b を読み取り、小さい方を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var a,b int; fmt.Scan(&a,&b); if a<b{ fmt.Println(a) } else { fmt.Println(b) } }\n```",
			Cases:     []testCase{{In: "3 5\n", Out: "3", Sample: true}, {In: "9 2\n", Out: "2", Sample: false}},
		},
	},
}
