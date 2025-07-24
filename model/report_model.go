package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportSummary merepresentasikan ringkasan laporan.
type ReportSummary struct {
	Period  string  `json:"period"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Profit  float64 `json:"profit"`
}

// ReportItem merepresentasikan item dalam kategori laporan.
type ReportItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

// ReportDetailCategory merepresentasikan kategori detail dalam laporan.
type ReportDetailCategory struct {
	Category string       `json:"category"`
	Items    []ReportItem `json:"items"`
}

// ReportData merepresentasikan data internal dari laporan.
type ReportData struct {
	Summary    ReportSummary          `json:"summary"`
	Details    []ReportDetailCategory `json:"details"`
	Conclusion string                 `json:"conclusion"`
}
type ReportVisualizationConfig struct {
	ReportType       string                 `json:"reportType"`             // e.g., "card", "bar_chart", "table"
	SourceDatabaseID primitive.ObjectID     `json:"sourceDatabaseId"`       // ID dari model Database yang akan di-query
	ChartConfig      map[string]interface{} `json:"chartConfig,omitempty"`  // Konfigurasi spesifik untuk chart (sumbu X, Y, dll)
	CardConfig       map[string]interface{} `json:"cardConfig,omitempty"`   // Konfigurasi spesifik untuk card (field yang ditampilkan)
	FilterConfig     map[string]interface{} `json:"filterConfig,omitempty"` // Konfigurasi filter data
	// ... bisa tambahkan field lain sesuai kebutuhan visualisasi
}

// Report merepresentasikan struktur dokumen laporan di database.
type Report struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChannelID primitive.ObjectID `bson:"channelId,omitempty" json:"channelId"`
	AuthorID  primitive.ObjectID `bson:"authorId,omitempty" json:"authorId"`
	Title     string             `json:"title"`
	// Ganti ReportData dengan konfigurasi visualisasi
	Configuration ReportVisualizationConfig `json:"configuration"`
	CreatedAt     time.Time                 `json:"createdAt"`
	UpdatedAt     time.Time                 `json:"updatedAt"`
}

/*
Cara Penggunaan:

// Contoh inisialisasi Report
// report := model.Report{
//     ID:        "report_123",
//     ChannelID: "channel_abc",
//     AuthorID:  "user_xyz",
//     Title:     "Laporan Bulanan",
//     ReportData: model.ReportData{
//         Summary: model.ReportSummary{
//             Period:  "Agustus 2024",
//             Income:  10000000,
//             Expense: 7000000,
//             Profit:  3000000,
//         },
//         Details: []model.ReportDetailCategory{
//             {
//                 Category: "Pemasukan",
//                 Items:    []model.ReportItem{{Description: "Penjualan", Amount: 10000000}},
//             },
//         },
//         Conclusion: "Laporan keuangan bulan ini stabil.",
//     },
//     CreatedAt: time.Now(),
//     UpdatedAt: time.Now(),
// }
*/
