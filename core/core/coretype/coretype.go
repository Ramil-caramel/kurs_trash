package coretype

// Пакет нужен для тогог чтобы разные пакет могли обмениваться общими типами
// Структура стандартизирующая хеш кусков

type Piece struct {
    Index  int
    Hash   [20]byte // SHA-1 (20 байт)
    Size   int
}