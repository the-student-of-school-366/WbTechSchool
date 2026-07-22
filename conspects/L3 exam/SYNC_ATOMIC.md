# Go Concurrency: Конспект для подготовки к экзамену L3


---

## 0. Карта темы (что вообще существует)

В Go concurrency строится на трёх слоях:

1. **Горутины + планировщик (GMP)** — лёгкие потоки исполнения.
2. **Каналы** — передача владения и синхронизация через коммуникацию
   (*share memory by communicating*).
3. **`sync` / `sync/atomic`** — классическая синхронизация над общей памятью
   (*share memory by protecting it*).

Пакет `sync` — высокоуровневые примитивы (блокировки, ожидание, пулы).
Пакет `sync/atomic` — низкоуровневые атомарные операции над одним словом
(или указателем/интерфейсом).

**Правило выбора:**
- Нужна передача данных / отмена / select → **канал**.
- Нужна защита структуры / инварианта из нескольких полей → **Mutex / RWMutex**.
- Счётчик / флаг / указатель на конфиг → **atomic**.
- Разовая инициализация → **Once / OnceValue**.
- Ожидание условия над состоянием под мьютексом → **Cond** (редко) или канал.
- Кэш объектов → **Pool** (осторожно).
- Concurrent map в узких сценариях → **sync.Map**.

---

## 1. `sync.Mutex`

### Зачем
Взаимоисключение: в критической секции одновременно работает **одна**
горутина. Гарантирует happens-before: всё, что сделано до `Unlock`, видно
тому, кто потом взял `Lock`.

```go
var mu sync.Mutex
var balance int

func Deposit(n int) {
    mu.Lock()
    defer mu.Unlock()
    balance += n
}
```

### Как устроен Mutex

```go
type Mutex struct {
    state int32  // битовый пакет
    sema  uint32 // семафор рантайма для парковки
}
```

Биты `state`:

| Бит | Имя | Смысл |
|-----|-----|-------|
| 0 | `mutexLocked` | лок занят |
| 1 | `mutexWoken` | кто-то уже разбужен / «лезет» на лок |
| 2 | `mutexStarving` | режим голодания (FIFO handoff) |
| 3… | waiter count | число ожидающих горутин |

**Fast path `Lock`:** `CAS(state, 0 → locked)`. Пустой мьютекс — одна
атомарная операция, без парковки.

**Slow path (`lockSlow`):**
1. На multi-CPU и не в starvation — **короткий spin** (`procyield` /
   `Gosched`): надежда, что лок освободится мгновенно.
2. Увеличиваем счётчик waiter’ов, паркуемся на `sema`
   (`runtime_SemacquireMutex`).
3. Если ждали **> ~1 мс** — включаем **starvation mode**: лок передаётся
   следующему waiter’у напрямую (handoff), без борьбы новых пришедших.
   Это FIFO и защита от голодания.
4. Выход из starvation — когда очередь пустеет или текущий waiter «молодой».

**`Unlock`:** снимает `mutexLocked`. Если есть waiter’ы — будит одного через
семафор. В starvation — handoff владельцу.

### Свойства (обязательно на экзамене)
- **Нереентерабельный:** повторный `Lock` той же горутиной → deadlock.
- `Unlock` без `Lock` → **panic**.
- Копировать нельзя (`vet` → `copylocks`; внутри есть методы `Lock`/`Unlock`).
- Нет публичного `IsLocked` — иначе провоцируют check-then-act.
- Спин + park + starvation: быстр при низком contention, честен при высоком.

### Почему не рекомендуется `TryLock`

`TryLock()` (Go 1.18) пытается взять лок без блокировки; `false`, если занят.
Дока буквально: *for instrumentation/debugging, not normal concurrency control*.

Почему плохо в обычном коде:
1. **Busy-wait:** `for !mu.TryLock() {}` жжёт CPU вместо парковки.
2. **Ломает fairness:** игнорирует очередь waiter’ов и starvation mode —
   может «украсть» лок у горутины, которая ждёт уже >1 мс.
3. **Livelock** при нескольких спиннерах на TryLock.
4. **Симптом плохого дизайна** (циклические зависимости под локом,
   reentrancy). Лучше: каналы, разнести данные, `select { default: }`.
5. `if mu.TryLock() { mu.Unlock() }` — бессмысленная гонка: ответ устаревает
   сразу.

Легитимно: метрики/дебаг («занят ли сейчас»), защита от reentrancy в
колбэке с понятным fallback.

### Что делает `noCopy` и почему первым

```go
type noCopy struct{}
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
```

Это **маркер для `go vet` (copylocks)**, а не реальный лок. Vet считает
типом-замком всё, у чего есть метод `Lock()`. Копирование `WaitGroup`,
`Cond`, `Mutex` (через встраивание/поля) ловится статически.

**Почему первым полем:**
1. **Zero-size aliasing.** Адрес zero-size поля в конце структуры может
   совпасть с адресом *следующей* аллокации. Первым — получает уникальный
   адрес (issue golang/go#8005).
2. **Выравнивание.** На 32-bit 64-битные атомики требуют 8-байтного
   выравнивания; первый word структуры гарантированно выровнен. `noCopy`
   (0 байт) первым позволяет поставить сразу за ним `atomic.Uint64`
   (пример — `WaitGroup`).

Порядок для самого `copylocks` не важен — vet обходит все поля. «Первым» —
про память и конвенцию.

### Вопросы / задачи

**В1.** Почему повторный `Lock` той же горутиной — deadlock, а не ошибка?
*Ответ:* Mutex не хранит владельца; реентерабельность усложнила бы рантайм и
маскировала бы архитектурные баги.

**В2.** Чем spinlock на `atomic.Bool` хуже `sync.Mutex`?
*Ответ:* нет парковки (жжёт CPU), нет starvation/FIFO, хуже под contention.

**Задача.** Найди баг:
```go
type Counter struct{ mu sync.Mutex; n int }
func (c Counter) Inc() { c.mu.Lock(); c.n++; c.mu.Unlock() }
```
*Ответ:* `Inc` принимает по значению → копирует Mutex → data race / бесполезный
лок. Нужен `(c *Counter)`.

---

## 2. `sync.RWMutex`

### Зачем
Много читателей одновременно **или** один писатель. Имеет смысл при
read-heavy и достаточно длинной секции чтения; иначе оверхед больше выгоды.

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]int
}

func (c *Cache) Get(k string) (int, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.items[k]
    return v, ok
}

func (c *Cache) Set(k string, v int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[k] = v
}
```

### Устройство (упрощённо)

```go
type RWMutex struct {
    w           Mutex
    writerSem   uint32
    readerSem   uint32
    readerCount atomic.Int32
    readerWait  atomic.Int32
}
const rwmutexMaxReaders = 1 << 30
```

- **`RLock`:** `readerCount.Add(1)`. Если ≥ 0 — ок (нет писателя). Если < 0 —
  писатель ждёт/пишет → парк на `readerSem`.
- **`RUnlock`:** `Add(-1)`; если был писатель и ты последний из «должных
  уйти» (`readerWait`) — будишь писателя.
- **`Lock` (write):** берёт `w` (сериализация писателей), затем
  `readerCount.Add(-rwmutexMaxReaders)` → отрицательный диапазон =
  «писатель хочет войти». Активные читатели дописываются в `readerWait`,
  писатель паркуется на `writerSem`.
- **`Unlock`:** возвращает `readerCount` в плюс, будит пачку читателей
  (`readerSem`, handoff).

### Нюансы
- Writer-preference: входящий писатель сразу делает `readerCount`
  отрицательным → новые `RLock` паркуются (частичная защита от starvation
  писателя).
- `RLock` + `Lock` одной горутиной на одном RWMutex → deadlock.
- Короткий `Get` (один map lookup) часто **быстрее** с обычным `Mutex` —
  бенчмаркать.
- `TryLock` / `TryRLock` — те же оговорки, что у Mutex.

### Вопросы
**В3.** Когда выбрать `Mutex`, а когда `RWMutex`?
*Ответ:* RWMutex — при большом read:write (≥ ~10:1) и длинном чтении.
Иначе Mutex проще и часто быстрее.

**В4.** Может ли писатель голодать при постоянном потоке читателей?
*Ответ:* В Go частично смягчено (новые читатели паркуются, когда писатель
ждёт), но при длинных RLock писатель всё равно может ждать долго.

---

## 3. `sync.WaitGroup`

### Зачем
Дождаться завершения набора горутин (fork-join).

```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(i int) {
        defer wg.Done()
        work(i)
    }(i)
}
wg.Wait()
```

### Устройство
```go
type WaitGroup struct {
    noCopy noCopy
    state  atomic.Uint64 // high 32: counter, low 32: waiters
    sema   uint32
}
```

- `Add(delta)` меняет counter; при уходе в ноль будит всех `Wait`.
- `Done()` = `Add(-1)`.
- `Wait` паркуется, пока counter ≠ 0.
- `Add` с уходом counter **ниже нуля** → panic.
- Копировать нельзя (`noCopy` первым + выравнивание `state`).

### Правила
1. `Add` **до** `go` (или `Add(n)` до цикла) — иначе `Wait` может увидеть 0
   раньше времени (data race на логике счётчика).
2. Не копировать WaitGroup после первого использования (передавать `*WaitGroup`).
3. Не вызывать `Add` положительный из «уже учтённых» горутин параллельно с
   `Wait` без аккуратности — классический источник багов.

### Вопросы
**В5.** Почему это плохо?
```go
for i := 0; i < n; i++ {
    go func() { wg.Add(1); defer wg.Done(); work() }()
}
wg.Wait()
```
*Ответ:* `Add` внутри горутины — гонка с `Wait`: Wait может стартовать при
counter=0.

**Задача.** Реализуй worker pool на N воркерах с graceful shutdown через
WaitGroup + context. Объясни, кто делает `Add`/`Done`.

---

## 4. `sync.Once`, `OnceValue`, `OnceFunc`, `OnceValues`

### `Once`
Гарантирует, что `f` выполнится **ровно один раз**. После `Do` все
вызывающие видят побочные эффекты `f` (happens-before).

```go
var (
    cfg  *Config
    once sync.Once
)

func GetConfig() *Config {
    once.Do(func() {
        cfg = load()
    })
    return cfg
}
```

**Устройство (идея):** atomic-флаг `done` + mutex на slow path
(double-checked locking **правильно**, потому что барьер даёт сам Once).

```go
func (o *Once) Do(f func()) {
    if o.done.Load() == 0 {
        o.doSlow(f) // под mutex + defer done.Store(1)
    }
}
```

Нюанс: если `f` **паникует**, `done` всё равно ставится (через `defer`) —
повторных попыток не будет. Для retry нужна другая схема.

### `OnceValue` / `OnceFunc` / `OnceValues` (Go 1.21+)

Синтаксический сахар над Once + результат:

```go
var getConfig = sync.OnceValue(func() *Config {
    return load()
})

cfg := getConfig() // потокобезопасно, один раз
```

| Хелпер | Сигнатура идеи | Когда |
|---|---|---|
| `OnceFunc(f)` | `func()` | side-effect once |
| `OnceValue(f)` | `func() T` | одно значение |
| `OnceValues(f)` | `func() (T1, T2)` | два значения (часто value, err) |

### Вопросы
**В6.** Чем Once лучше «ручного» double-checked locking с Mutex?
*Ответ:* Легко ошибиться в барьерах/памяти; Once уже правильный и с fast path.

**В7.** `OnceValue` паникует внутри — что будет при втором вызове?
*Ответ:* Паника «закэширована»: каждый следующий вызов снова паникует тем же
(Once уже done).

---

## 5. `sync.Cond`

### Что делает
**Condition variable:** «усни, пока условие не станет true», привязанная к
`Locker` (обычно `*Mutex` / `*RWMutex`).

```go
mu.Lock()
for !ready {
    cond.Wait() // атомарно: Unlock + park; при пробуждении снова Lock
}
// ready == true, mu удерживается
mu.Unlock()
```

Сигналящий:
```go
mu.Lock()
ready = true
cond.Broadcast() // или Signal()
mu.Unlock()
```

- `Wait` — отпустить лок и спать; проснуться и снова взять лок.
- `Signal` — разбудить **одного** waiter’а.
- `Broadcast` — **всех**.

### Когда использовать
- Несколько горутин ждут **разных условий** над одним защищённым состоянием.
- Нужен fan-out уведомления («конфиг обновился — все перечитайте»).
- Классические мониторы: bounded buffer, «ресурс освободился».

В идиоматичном Go **чаще каналы**. Cond — когда состояние уже под мьютексом
и канал неудобен.

### Проблемы Cond
1. **Нельзя `select`** — нет отмены по `ctx.Done()` / timeout из коробки.
2. **Потерянные сигналы:** `Signal` до `Wait` не запоминается (не канал).
   Условие-флаг обязательно; звонить лучше **под локом**.
3. **Spurious wakeups** → только `for !cond { Wait() }`, никогда `if`.
4. **Thundering herd** от `Broadcast`.
5. Копирование запрещено (`noCopy`), легко отстрелить ногу.
6. Сложнее читать/поддерживать, чем канал.

### Вопросы / задачи
**В8.** Почему цикл, а не `if` вокруг `Wait`?
*Ответ:* spurious wakeup + конкуренция после Broadcast (условие уже съели).

**Задача.** Напиши bounded buffer (Put/Get) на Mutex+Cond. Затем перепиши на
канале. Сравни: отмена по context, читаемость, производительность.

**В9.** Можно ли заменить Cond каналом буфера 1?
*Ответ:* Часто да для «событие случилось»; для сложных условий над общим
состоянием Cond компактнее, но канал + select гибче.

---

## 6. `sync.Pool`

### Зачем
Кэш **временных** объектов, чтобы снизить давление на GC (буферы,
`bytes.Buffer`, слайсы).

```go
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}

func handler() {
    b := bufPool.Get().(*bytes.Buffer)
    b.Reset()
    defer bufPool.Put(b)
    // ...
}
```

### Важные свойства
- **Нет гарантий:** `Get` может вернуть новый объект даже после `Put`.
- Между GC пул **может быть очищен** — нельзя хранить важное состояние.
- Локален per-P + victim cache — хорошо масштабируется, но семантика
  «мягкая».
- `New` вызывается, если взять нечего.
- Объекты должны быть **сбрасываемы** (`Reset`) перед Put — иначе утечка
  логики/памяти между запросами.

### Когда да / нет
- Да: горячие аллокации короткоживущих буферов одинакового размера.
- Нет: connection pool, кэш бизнес-данных, что-то с lifetime > одного GC.

### Вопросы
**В10.** Почему нельзя класть в Pool открытый `*sql.DB` / TCP conn?
*Ответ:* Pool может выкинуть объект при GC без `Close` → утечки FD.

**В11.** Что будет, если забыть `Reset` перед `Put`?
*Ответ:* Следующий `Get` получит «грязный» буфер — баги и раздувание памяти.

---

## 7. Обычная `map` vs concurrent доступ

Встроенная `map` **не потокобезопасна**. Параллельная запись или
write+read без синхронизации → data race; на практике часто fatal
«concurrent map writes».

Варианты:
| Подход | Когда |
|---|---|
| `map` + `Mutex` | write-heavy / простота |
| `map` + `RWMutex` | read-heavy, короткий/средний lookup |
| `atomic.Value` / `atomic.Pointer[map[K]V]` + COW | редкие обновления всей таблицы |
| `sync.Map` | узкие сценарии из доки (см. ниже) |
| шардирование map’ов | высокий contention на разных ключах |

---

## 8. `sync.Map` — подробно

### Когда использовать (из документации)
1. Ключ пишется **один раз**, читается много раз (растущий cache).
2. Много горутин работают с **разными** множествами ключей (disjoint keys).

Иначе обычно быстрее и проще `map + Mutex/RWMutex` (+ типобезопасность).

API: `Load`, `Store`, `Delete`, `LoadOrStore`, `LoadAndDelete`, `Swap`,
`CompareAndSwap`, `CompareAndDelete`, `Range`, `Clear` (1.23+).
Ключи/значения — `any` (исторически `interface{}`).

---

### Почему принимает `interface{}` / `any`?

1. **История:** `sync.Map` появился в Go 1.9 — **до дженериков** (1.18).
   Единственный способ «мапа от чего угодно» в публичном API — `interface{}`.
2. **Один тип на все кейсы:** внутри хранятся указатели/хеши единообразно.
3. **Обратная совместимость:** после 1.18 сигнатуру на `Map[K,V]` не меняли
   (сломало бы мир). Дженерики есть во **внутренней** реализации
   (`HashTrieMap[K,V]`), снаружи обёртка `HashTrieMap[any, any]`.
4. Цена: нет проверки типов на компиляции, аллокации при упаковке small
   values в interface, нужны type assert’ы.

Аналогичная история у `atomic.Value` (см. §10).

---

### Старая реализация (Go ≤ 1.23): `read` + `dirty`

Идея: **большинство чтений — без лока**.

```go
type Map struct {
    mu     Mutex
    read   atomic.Pointer[readOnly] // lock-free path
    dirty  map[any]*entry           // под mu
    misses int
}

type readOnly struct {
    m       map[any]*entry
    amended bool // dirty содержит ключи, которых нет в read
}

type entry struct {
    p atomic.Pointer[any] // *value | nil (удалён) | expunged
}
```

#### Что такое readOnly map?
«Чистый» снимок для **lock-free** чтения. Хранится за
`atomic.Pointer[readOnly]`. Горутина делает atomic load указателя и ищет
ключ в обычной Go-мапе **без мьютекса**.  
`amended == true` значит: есть ключи только в `dirty` — miss может уйти под
лок.

#### Что такое dirty map?
Полная мапа под `mu`: все актуальные ключи + новые, которых ещё нет в
`read`. Новые `Store` ключей идут сюда. Пока `dirty == nil`, при первой
необходимости её строят копированием не-expunged записей из `read`.

#### Состояния `entry.p`
- указатель на значение — живая запись;
- `nil` — логически удалено (в `read` ещё может торчать);
- `expunged` — удалено и **вычищено из dirty** при подготовке/промоушене:
  «в dirty этого ключа нет, в read — мусорный слот».

Зачем `expunged`: чтобы при создании `dirty` не тащить удалённый мусор и
отличать «можно снова вставить в dirty» от «просто nil».

#### Алгоритмы (старая)
- **Load:** смотрим `read` без лока. Miss и `amended` → под `mu` в `dirty`,
  `misses++`.
- **Store существующего:** часто CAS по `entry.p` без лока.
- **Store нового:** под `mu` в `dirty`.
- **Delete:** CAS в `nil`; при промоушене nil → expunged и не копируется.

#### Когда происходит promotion?
Когда число **miss’ов** (уходов в `dirty` из-за отсутствия в `read`)
превышает `len(dirty)` — считают, что `read` устарел. Под `mu`:
1. `dirty` становится новым `read` (`amended=false`);
2. `dirty = nil`;
3. `misses = 0`.

Цена: копирование/смена указателя → возможный **latency spike**. Удалённые
`expunged` живут в старом `read`, пока тот не отбросят — память может
держаться дольше, чем хочется.

**Слабые места старой:** глобальный `mu` на новые ключи; promotion spike;
expunged копятся; write-heavy плохо.

---

### Новая реализация (Go ≥ 1.24): HashTrieMap

```go
type Map struct {
    _ noCopy
    m isync.HashTrieMap[any, any]
}
```

`sync.Map` — тонкая обёртка. Внутри — **concurrent hash-trie**
(16-аричное дерево по нибблам хеша), не Swiss Table (Swiss Tables в 1.24 —
это встроенный `map`, другая история).

Идея:
- Корень — `atomic.Pointer` на узел.
- Внутренний узел: 16 детей + **локальный** `Mutex` на узел (не глобальный).
- Лист — `entry` с ключом/значением + overflow-цепочка при коллизиях.
- **Чтение** — lock-free спуск по хешу (4 бита уровня).
- **Запись** — оптимистичный поиск, затем лок **только нужного** узла,
  double-check, вставка / split при коллизии.
- **Удаление** + **compaction**: пустые узлы сжимаются вверх — нет вечного
  `expunged`-мусора.

#### Что изменилось vs старая
| | Старая | Новая (1.24+) |
|---|---|---|
| Структура | read + dirty + promotion | HashTrieMap |
| Лок | один глобальный `mu` | per-node mutex |
| Чтение hit | lock-free | lock-free |
| Новые ключи | всегда под глобальным mu | локальный узел |
| Promotion | да, по misses | **нет** (понятия read/dirty нет) |
| Удаления | expunged до promotion | compaction сразу |
| Write-heavy / mixed | слабо | заметно лучше |
| API | тот же | тот же |

Для экзамена важно уметь **обе** картинки: «классическую» (read/dirty —
её любят спрашивать) и «что в 1.24+».

---

### Вопросы по `sync.Map`

**В12.** Как устроен `sync.Map` (старый)?  
*Ключевые слова:* readOnly atomic, dirty под mu, entry+CAS, expunged, misses →
promotion.

**В13.** Что такое readOnly map?  
*Ответ:* lock-free снимок; `amended` сигналит, что dirty богаче.

**В14.** Что такое dirty map?  
*Ответ:* полная мапа под мьютексом для новых/промахнувшихся ключей.

**В15.** Когда promotion?  
*Ответ:* `misses > len(dirty)` → dirty становится новым read, dirty=nil.

**В16.** Почему `interface{}`?  
*Ответ:* до дженериков; совместимость; внутри 1.24 — generic HashTrieMap,
снаружи any.

**В17.** Что изменилось в новой реализации?  
*Ответ:* HashTrieMap, per-node locks, нет promotion/expunged, лучше запись и
память при удалениях.

**Задача.** Есть cache: 99% Load, редкий Store одного и того же набора ключей.
Что выбрать: `sync.Map`, `map+RWMutex`, `atomic.Pointer[map]`? Обоснуй и
придумай бенчмарк.

---

## 9. Пакет `sync/atomic` — обзор

Атомики дают:
- неделимость операции на одном слове/указателе;
- happens-before между atomic write и последующим atomic read той же
  переменной (в терминах Go memory model для sync/atomic).

**Не заменяют Mutex**, если инвариант из нескольких полей.

Философия: счётчики, флаги, публикация указателя на immutable-снимок.

---

## 10. Операции: Load, Store, Swap, CompareAndSwap, Add

На примере `atomic.Int64` (или устаревших `atomic.LoadInt64` и т.д.):

| Операция | Смысл |
|---|---|
| `Load` | атомарно прочитать |
| `Store` | атомарно записать |
| `Swap` | записать новое, вернуть старое |
| `CompareAndSwap(old, new)` | если сейчас `old` → поставить `new`, вернуть ok |
| `Add(delta)` | атомарно прибавить (для целых); вернуть новое |

```go
var n atomic.Int64
n.Store(10)
n.Add(1)                         // 11
old := n.Swap(0)                 // old=11, теперь 0
ok := n.CompareAndSwap(0, 42)    // true
```

**CAS** — основа lock-free циклов:
```go
for {
    old := p.Load()
    new := derive(old)
    if p.CompareAndSwap(old, new) {
        break
    }
}
```
Осторожно: ABA, спин под contention, сложность доказательства корректности.

---

## 11. `atomic.Value`

Хранит значение произвольного типа через `any`.

```go
var config atomic.Value // Store/Load any

config.Store(&Config{Addr: ":8080"})
cfg := config.Load().(*Config)
```

### Почему `interface{}`?
1. Добавлен **до дженериков** (публично стабилизирован раньше `Pointer[T]`).
2. Внутри — два слова интерфейса (type + data); атомарно меняется
   представление через CAS/Store указателей.
3. **Первый `Store` фиксирует тип**; другой тип → panic.
4. `Store(nil)` запрещён (panic). Можно хранить typed-nil `(*T)(nil)` после
   фиксации типа — нюанс для собеса.
5. После появления дженериков API не ломали; для указателей есть
   `atomic.Pointer[T]`.

### Когда использовать
Read-heavy публикация **целого** объекта (конфиг, таблица COW):
```go
var v atomic.Value
v.Store(map[string]int{"a": 1})

// обновление: копия + Store
old := v.Load().(map[string]int)
neu := maps.Clone(old)
neu["b"] = 2
v.Store(neu)
```
Читатели всегда видят согласованный снимок без лока.

---

## 12. `atomic.Pointer[T]` (Go 1.19+)

Типобезопасная замена частого паттерна «атомарный указатель».

```go
var p atomic.Pointer[Config]

p.Store(&Config{Addr: ":8080"})
cfg := p.Load() // *Config, без assert
```

Методы: `Load`, `Store`, `Swap`, `CompareAndSwap`.

vs `atomic.Value`:
- только указатели (или указатель-подобные), не произвольные values;
- нет «фиксации типа» через первый Store — тип в generics;
- обычно предпочтительнее для `*T`.

---

## 13. Новые generic atomic-типы (Go 1.19+)

Вместо `atomic.AddInt64(&x, 1)` — методы на значениях:

```go
var (
    flag atomic.Bool
    n    atomic.Int64
    u    atomic.Uint64
    up   atomic.Uintptr
)
flag.Store(true)
n.Add(1)
```

| Тип | Зачем |
|---|---|
| `Bool` | флаги |
| `Int32` / `Int64` | счётчики, эпохи |
| `Uint32` / `Uint64` | счётчики без знака |
| `Uintptr` | низкоуровневые трюки |
| `Pointer[T]` | публикация `*T` |
| `Value` | any / legacy / не-указатели |

Плюсы: нельзя передать «не тот» адрес, API яснее, меньше ошибок с
выравниванием (тип сам выровнен как поле структуры).

**Выравнивание:** на 32-bit `Int64`/`Uint64` должны быть 8-byte aligned —
поэтому их часто ставят первым полем (или сразу после `noCopy`).

---

## 14. Mutex ещё раз: шпаргалка «как отвечать»

1. **Устройство:** `state` (locked/woken/starving/waiters) + `sema`.
2. **Fast/slow path:** CAS → spin → park → starvation FIFO после ~1 мс.
3. **TryLock:** только дебаг/метрики; ломает fairness, провоцирует spin.
4. **noCopy:** vet-маркер; первым из-за zero-size aliasing и выравнивания.
5. **Не копировать, не реентерабелен, Unlock без Lock = panic.**

---

## 15. Cond ещё раз: шпаргалка «как отвечать»

1. **Что:** condition variable над Locker; Wait/Signal/Broadcast.
2. **Когда:** ждут изменения общего состояния; broadcast нескольким.
3. **Проблемы:** нет select/timeout, lost wakeups, spurious wakeups,
   thundering herd, сложнее каналов.
4. **Паттерн:** всегда `for !cond { Wait() }`, флаг под мьютексом, Signal
   под локом.

---

## 16. Сводная таблица выбора примитива

| Задача | Примитив |
|---|---|
| Критическая секция | `Mutex` |
| Много читателей | `RWMutex` (или `sync.Map` / COW) |
| Дождаться горутин | `WaitGroup` |
| Init once | `Once` / `OnceValue` |
| Ждать условие | канал или `Cond` |
| Переиспользовать буферы | `Pool` |
| Concurrent map (узко) | `sync.Map` |
| Счётчик / флаг | `atomic.Int64` / `Bool` |
| Опубликовать `*Config` | `atomic.Pointer[Config]` |
| Опубликовать any / COW map | `atomic.Value` |

---

## 17. Блок вопросов на понимание (мини-экзамен)

### Теория
1. В чём разница Mutex и RWMutex? Когда RWMutex проиграет Mutex?
2. Как WaitGroup гарантирует happens-before относительно `Wait`?
3. Почему `Once` безопаснее ручного double-checked locking?
4. Чем `OnceValue` отличается от глобальной `var x = f()`?
5. Зачем в Cond цикл, а не if?
6. Почему Pool очищается между GC и почему это OK для буферов?
7. Опиши read/dirty/expunged/promotion у старого sync.Map.
8. Что такое HashTrieMap и зачем per-node lock?
9. Почему sync.Map и atomic.Value принимают any?
10. Чем `atomic.Pointer[T]` лучше `atomic.Value` для `*T`?
11. Что делает CAS и где его недостаточно?
12. Как устроен Mutex (биты state, starvation)?
13. Почему TryLock не для production-логики?
14. Что такое noCopy и почему поле первым?
15. Три проблемы Cond, из-за которых в Go берут каналы.

### Практические задачки
**Z1. Safe counter**  
Реализуй `type Counter struct` с `Inc`/`Value` тремя способами: Mutex,
`atomic.Int64`, канал-актор. Сравни.

**Z2. Singleton config**  
Сломай и почини GetConfig: data race, Once, OnceValue, atomic.Pointer.

**Z3. Cache**  
Кэш с TTL: обоснованно выбери RWMutex vs sync.Map. Напиши бенчмарк
`Load` parallel.

**Z4. Worker pool**  
N воркеров, очередь задач, graceful shutdown: WaitGroup + context.
Где нельзя Cond?

**Z5. COW map**  
Словарь флагов фич: обновление раз в минуту, миллионы чтений/сек —
`atomic.Pointer[map[string]bool]`. Почему не sync.Map?

**Z6. Найди баг**
```go
var m sync.Map
m.Store(1, "a")
v, _ := m.Load(1)
fmt.Println(v + "b") // ?
```
*Ответ:* `v` имеет тип `any` — нельзя `+` без assert; будет ошибка компиляции.

**Z7. Найди баг**
```go
var once sync.Once
var err error
once.Do(func() { err = errors.New("fail") })
// хотим retry при err != nil
once.Do(func() { err = doInit() })
```
*Ответ:* второй Do no-op. Once не для retry; нужен ручной state или
singleflight + своя логика.

---

## Шпаргалка «что говорить»

| Тема | Ключевые слова |
|---|---|
| Mutex | state bits, spin+park, starvation ~1ms FIFO, non-reentrant |
| TryLock | instrumentation only, fairness break, busy-wait |
| noCopy | vet copylocks, zero-size aliasing, alignment |
| RWMutex | readerCount, writerSem, writer preference |
| WaitGroup | Add before go, counter\|waiters in Uint64, noCopy |
| Once / OnceValue | atomic done + mutex, happens-before, panic caches |
| Cond | Wait/Signal/Broadcast, for-loop, no select, lost wakeup |
| Pool | best-effort, GC flush, Reset before Put |
| sync.Map old | readOnly, dirty, expunged, misses→promotion |
| sync.Map new | HashTrieMap, per-node lock, no promotion |
| why any | pre-generics + compat |
| atomic ops | Load/Store/Swap/CAS/Add |
| Pointer vs Value | typed *T vs any + first-store type |
| generic atomics | Bool/Int64/… methods, alignment |

---

*Опирается на pkg.go.dev/sync, pkg.go.dev/sync/atomic, исходники
`sync/mutex.go`, `sync/rwmutex.go`, `sync/map.go`, `internal/sync` (HashTrieMap),
go.dev/blog (Go 1.24 maps / sync.Map), и разделы из `INTERVIEW_CONCURRENCY.md`.*
