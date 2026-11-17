package domain

// Language = satu pilihan di layar "Choose board language".
type Language struct {
	Code        string `json:"code"`         // kode yang disimpan & dikirim, mis. "en"
	Name        string `json:"name"`         // nama dalam bahasanya sendiri, mis. "中文"
	EnglishName string `json:"english_name"` // keterangan kecil di bawahnya, mis. "Mandarin"
}

// DefaultLanguage dipakai kalau guru belum pernah memilih.
const DefaultLanguage = "id"

// Languages = bahasa yang didukung papan, urut seperti di layar.
// Kalau menambah bahasa di sini, tambahkan juga kodenya di CHECK constraint
// migrations/003_user_language.sql.
var Languages = []Language{
	{Code: "id", Name: "Bahasa Indonesia", EnglishName: "Indonesian"},
	{Code: "en", Name: "English", EnglishName: "English"},
	{Code: "zh", Name: "中文", EnglishName: "Mandarin"},
	{Code: "es", Name: "Español", EnglishName: "Spanish"},
	{Code: "ar", Name: "العربية", EnglishName: "Arabic"},
}
