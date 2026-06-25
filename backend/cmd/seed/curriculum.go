package main

import (
	"encoding/json"
	"strings"
)

// ml is a language→text map. get falls back to English when a language is missing.
type ml map[string]string

func (m ml) get(lang string) string {
	if v, ok := m[lang]; ok && strings.TrimSpace(v) != "" {
		return v
	}
	return m["en"]
}

type option struct {
	Text    ml
	Correct bool
}
type question struct {
	Prompt  ml
	Type    string // "single" | "multiple"
	Options []option
}
type testCase struct {
	In     string
	Out    string
	Sample bool
}
type problemSpec struct {
	Slug      string
	Title     ml
	Statement ml
	Solution  string // markdown (shared; code is language-agnostic)
	Cases     []testCase
}

func (p problemSpec) sampleJSON() string {
	type io struct {
		Input  string `json:"input"`
		Output string `json:"output"`
	}
	out := make([]io, 0)
	for _, c := range p.Cases {
		if c.Sample {
			out = append(out, io{c.In, c.Out})
		}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

type topic struct {
	Tag        string
	Difficulty string
	Title      ml
	Blurb      ml
	Article    ml
	Quiz       []question
	Prob       problemSpec
}

type project struct {
	Tag        string
	Difficulty string
	Title      ml
	Desc       ml
	Steps      []ml
}

// single-correct question helper
func sq(prompt ml, opts ...option) question {
	return question{Prompt: prompt, Type: "single", Options: opts}
}
func opt(text ml, correct bool) option { return option{Text: text, Correct: correct} }

var curriculum = []topic{
	{
		Tag: "go-hello", Difficulty: "beginner",
		Title: ml{"ru": "Привет, мир", "en": "Hello, World", "uz": "Salom, dunyo", "ja": "Hello, World"},
		Blurb: ml{"ru": "Первая программа на Go, пакеты и go run.", "en": "Your first Go program, packages and go run.", "uz": "Birinchi Go dasturi, paketlar va go run.", "ja": "最初の Go プログラム、パッケージと go run。"},
		Article: ml{
			"en": "# Hello, World\n\nEvery Go program starts in package `main` with a `main` function. Run it with `go run`.\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World\")\n}\n```",
			"ru": "# Привет, мир\n\nЛюбая программа на Go начинается с пакета `main` и функции `main`. Запуск — `go run`.\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Привет, мир\")\n}\n```",
			"uz": "# Salom, dunyo\n\nHar bir Go dasturi `main` paketidan va `main` funksiyasidan boshlanadi. `go run` bilan ishga tushiriladi.\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Salom, dunyo\")\n}\n```",
			"ja": "# Hello, World\n\nGo のプログラムは `main` パッケージの `main` 関数から始まります。`go run` で実行します。\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World\")\n}\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Какой пакет обязателен для исполняемой программы?", "en": "Which package is required for an executable program?", "uz": "Bajariladigan dastur uchun qaysi paket talab qilinadi?", "ja": "実行可能プログラムに必要なパッケージは？"},
				opt(ml{"ru": "main", "en": "main", "uz": "main", "ja": "main"}, true),
				opt(ml{"ru": "fmt", "en": "fmt", "uz": "fmt", "ja": "fmt"}, false),
				opt(ml{"ru": "exec", "en": "exec", "uz": "exec", "ja": "exec"}, false)),
			sq(ml{"ru": "Чем запускают программу без сборки бинаря?", "en": "Which command runs a program without building a binary?", "uz": "Binarsiz dasturni ishga tushiradigan buyruq?", "ja": "バイナリを作らずに実行するコマンドは？"},
				opt(ml{"ru": "go run", "en": "go run", "uz": "go run", "ja": "go run"}, true),
				opt(ml{"ru": "go vet", "en": "go vet", "uz": "go vet", "ja": "go vet"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-echo", Title: ml{"ru": "Эхо", "en": "Echo", "uz": "Aks-sado", "ja": "エコー"},
			Statement: ml{"ru": "Прочитайте строку и выведите её без изменений.", "en": "Read a line and print it unchanged.", "uz": "Bir qatorni o'qing va o'zgartirmasdan chiqaring.", "ja": "1 行を読み取り、そのまま出力してください。"},
			Solution:  "```go\npackage main\nimport (\"bufio\";\"fmt\";\"os\")\nfunc main(){ r:=bufio.NewReader(os.Stdin); s,_:=r.ReadString('\\n'); fmt.Print(s) }\n```",
			Cases:     []testCase{{In: "hello\n", Out: "hello", Sample: true}, {In: "go\n", Out: "go", Sample: false}},
		},
	},
	{
		Tag: "go-types", Difficulty: "beginner",
		Title: ml{"ru": "Переменные и типы", "en": "Variables and types", "uz": "Oʻzgaruvchilar va turlar", "ja": "変数と型"},
		Blurb: ml{"ru": "Объявление переменных, базовые типы, константы.", "en": "Declaring variables, basic types, constants.", "uz": "Oʻzgaruvchilarni e'lon qilish, asosiy turlar, konstantalar.", "ja": "変数の宣言、基本型、定数。"},
		Article: ml{
			"en": "# Variables and types\n\nUse `var` or short `:=`. Go has `int`, `float64`, `string`, `bool`. `const` declares constants.\n\n```go\nvar n int = 42\nname := \"Go\"\nconst Pi = 3.14159\n```",
			"ru": "# Переменные и типы\n\nИспользуйте `var` или короткое `:=`. Базовые типы: `int`, `float64`, `string`, `bool`. `const` — константы.\n\n```go\nvar n int = 42\nname := \"Go\"\nconst Pi = 3.14159\n```",
			"uz": "# Oʻzgaruvchilar va turlar\n\n`var` yoki qisqa `:=` ishlating. Asosiy turlar: `int`, `float64`, `string`, `bool`. `const` — konstanta.\n\n```go\nvar n int = 42\nname := \"Go\"\nconst Pi = 3.14159\n```",
			"ja": "# 変数と型\n\n`var` または短縮形 `:=` を使います。基本型は `int`、`float64`、`string`、`bool`。定数は `const`。\n\n```go\nvar n int = 42\nname := \"Go\"\nconst Pi = 3.14159\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что объявляет короткое `:=`?", "en": "What does the short `:=` declare?", "uz": "Qisqa `:=` nima e'lon qiladi?", "ja": "短縮形 `:=` は何を宣言する？"},
				opt(ml{"ru": "Переменную с выводом типа", "en": "A variable with inferred type", "uz": "Turi aniqlangan oʻzgaruvchi", "ja": "型推論された変数"}, true),
				opt(ml{"ru": "Только константу", "en": "A constant only", "uz": "Faqat konstanta", "ja": "定数のみ"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-sum", Title: ml{"ru": "Сумма двух чисел", "en": "Sum of two numbers", "uz": "Ikki sonning yigʻindisi", "ja": "2 つの数の和"},
			Statement: ml{"ru": "Прочитайте два целых числа и выведите их сумму.", "en": "Read two integers and print their sum.", "uz": "Ikkita butun sonni o'qing va yig'indisini chiqaring.", "ja": "2 つの整数を読み取り、その和を出力してください。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var a,b int; fmt.Scan(&a,&b); fmt.Println(a+b) }\n```",
			Cases:     []testCase{{In: "2 3\n", Out: "5", Sample: true}, {In: "10 20\n", Out: "30", Sample: false}, {In: "-1 1\n", Out: "0", Sample: false}},
		},
	},
	{
		Tag: "go-flow", Difficulty: "beginner",
		Title: ml{"ru": "Управление потоком", "en": "Control flow", "uz": "Boshqaruv oqimi", "ja": "制御フロー"},
		Blurb: ml{"ru": "if, for и switch в Go.", "en": "if, for and switch in Go.", "uz": "Go'da if, for va switch.", "ja": "Go の if、for、switch。"},
		Article: ml{
			"en": "# Control flow\n\nGo has only one loop keyword: `for`. `if` can carry an init statement; `switch` needs no `break`.\n\n```go\nfor i := 0; i < 3; i++ {\n\tif i%2 == 0 {\n\t\tfmt.Println(i)\n\t}\n}\n```",
			"ru": "# Управление потоком\n\nВ Go один цикл — `for`. `if` может иметь init-выражение; в `switch` не нужен `break`.\n\n```go\nfor i := 0; i < 3; i++ {\n\tif i%2 == 0 {\n\t\tfmt.Println(i)\n\t}\n}\n```",
			"uz": "# Boshqaruv oqimi\n\nGo'da bitta sikl bor: `for`. `if` init-ifodasi bilan bo'lishi mumkin; `switch`'da `break` shart emas.\n\n```go\nfor i := 0; i < 3; i++ {\n\tif i%2 == 0 {\n\t\tfmt.Println(i)\n\t}\n}\n```",
			"ja": "# 制御フロー\n\nGo のループは `for` だけです。`if` は初期化文を持てます。`switch` に `break` は不要です。\n\n```go\nfor i := 0; i < 3; i++ {\n\tif i%2 == 0 {\n\t\tfmt.Println(i)\n\t}\n}\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Сколько ключевых слов для циклов в Go?", "en": "How many loop keywords does Go have?", "uz": "Go'da nechta sikl kalit so'zi bor?", "ja": "Go のループキーワードはいくつ？"},
				opt(ml{"ru": "Одно: for", "en": "One: for", "uz": "Bitta: for", "ja": "1 つ: for"}, true),
				opt(ml{"ru": "Три: for, while, do", "en": "Three: for, while, do", "uz": "Uchta: for, while, do", "ja": "3 つ: for, while, do"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-fizzbuzz", Title: ml{"ru": "FizzBuzz", "en": "FizzBuzz", "uz": "FizzBuzz", "ja": "FizzBuzz"},
			Statement: ml{"ru": "Для n выведите числа 1..n; кратные 3 → Fizz, 5 → Buzz, 15 → FizzBuzz, каждое с новой строки.", "en": "Given n, print 1..n; multiples of 3 → Fizz, 5 → Buzz, 15 → FizzBuzz, one per line.", "uz": "n uchun 1..n sonlarni chiqaring; 3 ga → Fizz, 5 ga → Buzz, 15 ga → FizzBuzz, har biri yangi qatordan.", "ja": "n が与えられたとき 1..n を出力。3 の倍数→Fizz、5→Buzz、15→FizzBuzz、1 行ずつ。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var n int; fmt.Scan(&n); for i:=1;i<=n;i++{ switch{ case i%15==0: fmt.Println(\"FizzBuzz\"); case i%3==0: fmt.Println(\"Fizz\"); case i%5==0: fmt.Println(\"Buzz\"); default: fmt.Println(i) } } }\n```",
			Cases:     []testCase{{In: "5\n", Out: "1\n2\nFizz\n4\nBuzz", Sample: true}, {In: "15\n", Out: "1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz", Sample: false}},
		},
	},
	{
		Tag: "go-funcs", Difficulty: "beginner",
		Title: ml{"ru": "Функции", "en": "Functions", "uz": "Funksiyalar", "ja": "関数"},
		Blurb: ml{"ru": "Параметры, множественный возврат, ошибки.", "en": "Parameters, multiple returns, errors.", "uz": "Parametrlar, koʻp qiymat qaytarish, xatolar.", "ja": "引数、複数戻り値、エラー。"},
		Article: ml{
			"en": "# Functions\n\nFunctions can return multiple values — idiomatically `(result, error)`.\n\n```go\nfunc div(a, b int) (int, error) {\n\tif b == 0 {\n\t\treturn 0, fmt.Errorf(\"division by zero\")\n\t}\n\treturn a / b, nil\n}\n```",
			"ru": "# Функции\n\nФункции возвращают несколько значений — идиома `(result, error)`.\n\n```go\nfunc div(a, b int) (int, error) {\n\tif b == 0 {\n\t\treturn 0, fmt.Errorf(\"деление на ноль\")\n\t}\n\treturn a / b, nil\n}\n```",
			"uz": "# Funksiyalar\n\nFunksiyalar bir nechta qiymat qaytaradi — odatda `(result, error)`.\n\n```go\nfunc div(a, b int) (int, error) {\n\tif b == 0 {\n\t\treturn 0, fmt.Errorf(\"nolga bo'lish\")\n\t}\n\treturn a / b, nil\n}\n```",
			"ja": "# 関数\n\n関数は複数の値を返せます。慣用的には `(result, error)`。\n\n```go\nfunc div(a, b int) (int, error) {\n\tif b == 0 {\n\t\treturn 0, fmt.Errorf(\"ゼロ除算\")\n\t}\n\treturn a / b, nil\n}\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что обычно возвращают функции Go вторым значением?", "en": "What do Go functions usually return as the second value?", "uz": "Go funksiyalari ikkinchi qiymat sifatida odatda nima qaytaradi?", "ja": "Go の関数が 2 番目の戻り値として通常返すものは？"},
				opt(ml{"ru": "error", "en": "error", "uz": "error", "ja": "error"}, true),
				opt(ml{"ru": "bool", "en": "bool", "uz": "bool", "ja": "bool"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-max", Title: ml{"ru": "Максимум", "en": "Maximum", "uz": "Maksimum", "ja": "最大値"},
			Statement: ml{"ru": "Прочитайте n и n чисел, выведите максимум.", "en": "Read n then n integers, print the maximum.", "uz": "n va n ta sonni o'qing, maksimumni chiqaring.", "ja": "n と n 個の整数を読み取り、最大値を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var n int; fmt.Scan(&n); m:=-1<<62; for i:=0;i<n;i++{ var x int; fmt.Scan(&x); if x>m{m=x} }; fmt.Println(m) }\n```",
			Cases:     []testCase{{In: "3\n1 5 2\n", Out: "5", Sample: true}, {In: "4\n-3 -1 -9 -2\n", Out: "-1", Sample: false}},
		},
	},
	{
		Tag: "go-collections", Difficulty: "intermediate",
		Title: ml{"ru": "Срезы и карты", "en": "Slices and maps", "uz": "Slice va map", "ja": "スライスとマップ"},
		Blurb: ml{"ru": "Динамические массивы и ассоциативные массивы.", "en": "Dynamic arrays and associative arrays.", "uz": "Dinamik massivlar va assotsiativ massivlar.", "ja": "動的配列と連想配列。"},
		Article: ml{
			"en": "# Slices and maps\n\n`[]T` is a growable slice; `map[K]V` is a hash map.\n\n```go\ns := []int{1, 2}\ns = append(s, 3)\nm := map[string]int{\"a\": 1}\nm[\"b\"] = 2\n```",
			"ru": "# Срезы и карты\n\n`[]T` — растущий срез; `map[K]V` — хеш-таблица.\n\n```go\ns := []int{1, 2}\ns = append(s, 3)\nm := map[string]int{\"a\": 1}\nm[\"b\"] = 2\n```",
			"uz": "# Slice va map\n\n`[]T` — kengayadigan slice; `map[K]V` — hash jadval.\n\n```go\ns := []int{1, 2}\ns = append(s, 3)\nm := map[string]int{\"a\": 1}\nm[\"b\"] = 2\n```",
			"ja": "# スライスとマップ\n\n`[]T` は伸長可能なスライス、`map[K]V` はハッシュマップです。\n\n```go\ns := []int{1, 2}\ns = append(s, 3)\nm := map[string]int{\"a\": 1}\nm[\"b\"] = 2\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Чем добавляют элемент в срез?", "en": "How do you add an element to a slice?", "uz": "Slice'ga element qanday qo'shiladi?", "ja": "スライスに要素を追加するには？"},
				opt(ml{"ru": "append", "en": "append", "uz": "append", "ja": "append"}, true),
				opt(ml{"ru": "push", "en": "push", "uz": "push", "ja": "push"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-unique", Title: ml{"ru": "Уникальные числа", "en": "Unique count", "uz": "Noyob sonlar", "ja": "ユニーク数"},
			Statement: ml{"ru": "Прочитайте n и n чисел, выведите количество различных значений.", "en": "Read n then n integers, print how many distinct values.", "uz": "n va n ta sonni o'qing, nechta turli qiymat borligini chiqaring.", "ja": "n と n 個の整数を読み取り、異なる値の個数を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var n int; fmt.Scan(&n); set:=map[int]bool{}; for i:=0;i<n;i++{ var x int; fmt.Scan(&x); set[x]=true }; fmt.Println(len(set)) }\n```",
			Cases:     []testCase{{In: "5\n1 2 2 3 3\n", Out: "3", Sample: true}, {In: "3\n7 7 7\n", Out: "1", Sample: false}},
		},
	},
	{
		Tag: "go-structs", Difficulty: "intermediate",
		Title: ml{"ru": "Структуры и методы", "en": "Structs and methods", "uz": "Struct va metodlar", "ja": "構造体とメソッド"},
		Blurb: ml{"ru": "Свои типы данных и методы на них.", "en": "Your own data types and methods on them.", "uz": "O'z ma'lumot turlaringiz va ularning metodlari.", "ja": "独自のデータ型とそのメソッド。"},
		Article: ml{
			"en": "# Structs and methods\n\nA `struct` groups fields; methods attach behaviour to a type.\n\n```go\ntype Point struct{ X, Y int }\nfunc (p Point) Sum() int { return p.X + p.Y }\n```",
			"ru": "# Структуры и методы\n\n`struct` группирует поля; методы добавляют поведение типу.\n\n```go\ntype Point struct{ X, Y int }\nfunc (p Point) Sum() int { return p.X + p.Y }\n```",
			"uz": "# Struct va metodlar\n\n`struct` maydonlarni guruhlaydi; metodlar turga xatti-harakat qo'shadi.\n\n```go\ntype Point struct{ X, Y int }\nfunc (p Point) Sum() int { return p.X + p.Y }\n```",
			"ja": "# 構造体とメソッド\n\n`struct` はフィールドをまとめ、メソッドは型に振る舞いを付与します。\n\n```go\ntype Point struct{ X, Y int }\nfunc (p Point) Sum() int { return p.X + p.Y }\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Что группирует именованные поля?", "en": "What groups named fields together?", "uz": "Nomli maydonlarni nima guruhlaydi?", "ja": "名前付きフィールドをまとめるのは？"},
				opt(ml{"ru": "struct", "en": "struct", "uz": "struct", "ja": "struct"}, true),
				opt(ml{"ru": "class", "en": "class", "uz": "class", "ja": "class"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-rectangle", Title: ml{"ru": "Площадь прямоугольника", "en": "Rectangle area", "uz": "To'rtburchak yuzasi", "ja": "長方形の面積"},
			Statement: ml{"ru": "Прочитайте ширину и высоту, выведите площадь.", "en": "Read width and height, print the area.", "uz": "Eni va bo'yini o'qing, yuzasini chiqaring.", "ja": "幅と高さを読み取り、面積を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var w,h int; fmt.Scan(&w,&h); fmt.Println(w*h) }\n```",
			Cases:     []testCase{{In: "3 4\n", Out: "12", Sample: true}, {In: "10 10\n", Out: "100", Sample: false}},
		},
	},
	{
		Tag: "go-interfaces", Difficulty: "intermediate",
		Title: ml{"ru": "Интерфейсы", "en": "Interfaces", "uz": "Interfeyslar", "ja": "インターフェース"},
		Blurb: ml{"ru": "Полиморфизм через поведение.", "en": "Polymorphism through behaviour.", "uz": "Xatti-harakat orqali polimorfizm.", "ja": "振る舞いによる多態性。"},
		Article: ml{
			"en": "# Interfaces\n\nAn interface is a set of method signatures; types satisfy it implicitly.\n\n```go\ntype Stringer interface{ String() string }\n```",
			"ru": "# Интерфейсы\n\nИнтерфейс — набор сигнатур методов; типы удовлетворяют ему неявно.\n\n```go\ntype Stringer interface{ String() string }\n```",
			"uz": "# Interfeyslar\n\nInterfeys — metod imzolari to'plami; turlar unga oshkormas ravishda mos keladi.\n\n```go\ntype Stringer interface{ String() string }\n```",
			"ja": "# インターフェース\n\nインターフェースはメソッドシグネチャの集合で、型は暗黙的に満たします。\n\n```go\ntype Stringer interface{ String() string }\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Как тип реализует интерфейс в Go?", "en": "How does a type implement an interface in Go?", "uz": "Go'da tur interfeysni qanday amalga oshiradi?", "ja": "Go で型はどのようにインターフェースを実装する？"},
				opt(ml{"ru": "Неявно, имея нужные методы", "en": "Implicitly, by having the methods", "uz": "Oshkormas — kerakli metodlarga ega bo'lib", "ja": "暗黙的に、メソッドを持つことで"}, true),
				opt(ml{"ru": "Через ключевое слово implements", "en": "With an implements keyword", "uz": "implements kalit so'zi bilan", "ja": "implements キーワードで"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-abs", Title: ml{"ru": "Модуль числа", "en": "Absolute value", "uz": "Absolyut qiymat", "ja": "絶対値"},
			Statement: ml{"ru": "Прочитайте целое число и выведите его модуль.", "en": "Read an integer and print its absolute value.", "uz": "Butun sonni o'qing va absolyut qiymatini chiqaring.", "ja": "整数を読み取り、その絶対値を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var x int; fmt.Scan(&x); if x<0{x=-x}; fmt.Println(x) }\n```",
			Cases:     []testCase{{In: "-7\n", Out: "7", Sample: true}, {In: "5\n", Out: "5", Sample: false}},
		},
	},
	{
		Tag: "go-concurrency", Difficulty: "advanced",
		Title: ml{"ru": "Горутины и каналы", "en": "Goroutines and channels", "uz": "Goroutine va kanallar", "ja": "ゴルーチンとチャネル"},
		Blurb: ml{"ru": "Конкурентность в Go.", "en": "Concurrency in Go.", "uz": "Go'da parallellik.", "ja": "Go の並行処理。"},
		Article: ml{
			"en": "# Goroutines and channels\n\nStart a goroutine with `go`; communicate over a typed channel.\n\n```go\nch := make(chan int)\ngo func() { ch <- 42 }()\nfmt.Println(<-ch)\n```",
			"ru": "# Горутины и каналы\n\nЗапуск горутины — `go`; обмен данными — через типизированный канал.\n\n```go\nch := make(chan int)\ngo func() { ch <- 42 }()\nfmt.Println(<-ch)\n```",
			"uz": "# Goroutine va kanallar\n\nGoroutine'ni `go` bilan ishga tushiring; ma'lumot almashish — kanal orqali.\n\n```go\nch := make(chan int)\ngo func() { ch <- 42 }()\nfmt.Println(<-ch)\n```",
			"ja": "# ゴルーチンとチャネル\n\n`go` でゴルーチンを起動し、型付きチャネルで通信します。\n\n```go\nch := make(chan int)\ngo func() { ch <- 42 }()\nfmt.Println(<-ch)\n```",
		},
		Quiz: []question{
			sq(ml{"ru": "Каким словом запускают горутину?", "en": "Which keyword starts a goroutine?", "uz": "Goroutine qaysi kalit so'z bilan ishga tushadi?", "ja": "ゴルーチンを起動するキーワードは？"},
				opt(ml{"ru": "go", "en": "go", "uz": "go", "ja": "go"}, true),
				opt(ml{"ru": "async", "en": "async", "uz": "async", "ja": "async"}, false)),
		},
		Prob: problemSpec{
			Slug: "go-sumn", Title: ml{"ru": "Сумма последовательности", "en": "Sum 1..n", "uz": "1..n yig'indisi", "ja": "1..n の和"},
			Statement: ml{"ru": "Прочитайте n и выведите сумму 1+2+...+n.", "en": "Read n and print 1+2+...+n.", "uz": "n ni o'qing va 1+2+...+n yig'indisini chiqaring.", "ja": "n を読み取り 1+2+...+n を出力。"},
			Solution:  "```go\npackage main\nimport \"fmt\"\nfunc main(){ var n int; fmt.Scan(&n); fmt.Println(n*(n+1)/2) }\n```",
			Cases:     []testCase{{In: "5\n", Out: "15", Sample: true}, {In: "100\n", Out: "5050", Sample: false}},
		},
	},
}

var projects = []project{
	{
		Tag: "proj-cli-todo", Difficulty: "beginner",
		Title: ml{"ru": "CLI: список дел", "en": "CLI to-do app", "uz": "CLI vazifalar ro'yxati", "ja": "CLI ToDo アプリ"},
		Desc:  ml{"ru": "# Список дел\n\nКонсольное приложение для управления задачами.", "en": "# To-do app\n\nA command-line app to manage tasks.", "uz": "# Vazifalar ro'yxati\n\nVazifalarni boshqaruvchi konsol ilovasi.", "ja": "# ToDo アプリ\n\nタスクを管理するコマンドラインアプリ。"},
		Steps: []ml{
			{"ru": "Разобрать аргументы командной строки", "en": "Parse command-line arguments", "uz": "Buyruq qatori argumentlarini tahlil qilish", "ja": "コマンドライン引数を解析する"},
			{"ru": "Хранить задачи в JSON-файле", "en": "Persist tasks to a JSON file", "uz": "Vazifalarni JSON faylga saqlash", "ja": "タスクを JSON ファイルに保存する"},
			{"ru": "Команды add / list / done", "en": "add / list / done commands", "uz": "add / list / done buyruqlari", "ja": "add / list / done コマンド"},
		},
	},
	{
		Tag: "proj-http-api", Difficulty: "intermediate",
		Title: ml{"ru": "HTTP JSON API", "en": "HTTP JSON API", "uz": "HTTP JSON API", "ja": "HTTP JSON API"},
		Desc:  ml{"ru": "# HTTP API\n\nМаленький REST-сервис на net/http.", "en": "# HTTP API\n\nA small REST service with net/http.", "uz": "# HTTP API\n\nnet/http bilan kichik REST xizmat.", "ja": "# HTTP API\n\nnet/http による小さな REST サービス。"},
		Steps: []ml{
			{"ru": "Поднять сервер на net/http", "en": "Start a net/http server", "uz": "net/http serverini ishga tushirish", "ja": "net/http サーバーを起動する"},
			{"ru": "Маршруты и JSON-ответы", "en": "Routes and JSON responses", "uz": "Marshrutlar va JSON javoblar", "ja": "ルートと JSON レスポンス"},
			{"ru": "Обработка ошибок", "en": "Error handling", "uz": "Xatolarni qayta ishlash", "ja": "エラー処理"},
		},
	},
	{
		Tag: "proj-wc", Difficulty: "beginner",
		Title: ml{"ru": "Счётчик слов", "en": "Word counter", "uz": "So'z sanagich", "ja": "単語カウンター"},
		Desc:  ml{"ru": "# Счётчик слов\n\nЧитает текст и считает слова/строки.", "en": "# Word counter\n\nReads text and counts words/lines.", "uz": "# So'z sanagich\n\nMatnni o'qiydi va so'z/qatorlarni sanaydi.", "ja": "# 単語カウンター\n\nテキストを読み、単語数・行数を数える。"},
		Steps: []ml{
			{"ru": "Читать stdin построчно", "en": "Read stdin line by line", "uz": "stdin'ni qatorma-qator o'qish", "ja": "stdin を 1 行ずつ読む"},
			{"ru": "Считать слова и строки", "en": "Count words and lines", "uz": "So'z va qatorlarni sanash", "ja": "単語と行を数える"},
		},
	},
	{
		Tag: "proj-guess", Difficulty: "beginner",
		Title: ml{"ru": "Игра «Угадай число»", "en": "Guess-the-number game", "uz": "Sonni toping o'yini", "ja": "数当てゲーム"},
		Desc:  ml{"ru": "# Угадай число\n\nКонсольная игра с генерацией случайного числа.", "en": "# Guess the number\n\nA console game with a random number.", "uz": "# Sonni toping\n\nTasodifiy son bilan konsol o'yini.", "ja": "# 数当て\n\n乱数を使うコンソールゲーム。"},
		Steps: []ml{
			{"ru": "Сгенерировать случайное число", "en": "Generate a random number", "uz": "Tasodifiy son yaratish", "ja": "乱数を生成する"},
			{"ru": "Цикл ввода и подсказки", "en": "Input loop with hints", "uz": "Kiritish sikli va maslahatlar", "ja": "入力ループとヒント"},
		},
	},
	{
		Tag: "proj-url-shortener", Difficulty: "advanced",
		Title: ml{"ru": "Сократитель ссылок", "en": "URL shortener", "uz": "URL qisqartirgich", "ja": "URL 短縮ツール"},
		Desc:  ml{"ru": "# Сократитель ссылок\n\nСервис с хранением и редиректами.", "en": "# URL shortener\n\nA service with storage and redirects.", "uz": "# URL qisqartirgich\n\nSaqlash va yo'naltirishli xizmat.", "ja": "# URL 短縮\n\nストレージとリダイレクトを持つサービス。"},
		Steps: []ml{
			{"ru": "Генерация коротких кодов", "en": "Generate short codes", "uz": "Qisqa kodlar yaratish", "ja": "短いコードを生成する"},
			{"ru": "Хранилище код→URL", "en": "code→URL storage", "uz": "kod→URL saqlash", "ja": "コード→URL ストレージ"},
			{"ru": "HTTP-редирект", "en": "HTTP redirect", "uz": "HTTP yo'naltirish", "ja": "HTTP リダイレクト"},
		},
	},
}
