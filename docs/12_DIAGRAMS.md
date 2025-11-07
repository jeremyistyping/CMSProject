# Diagram Flow (Mermaid)

Berikut contoh diagram untuk membantu visualisasi alur. Anda dapat memperluas sesuai proses bisnis.

## Sales → Journal → Reports
```mermaid
flowchart LR
  A[Create Sale] --> B[Confirm / Invoice]
  B --> C[Record Payment]
  B --> D[Generate Journal: AR/Revenue/PPN]
  C --> E[Generate Journal: Bank/AR]
  D --> F[Post Journals]
  F --> G[Reports: P&L, AR Aging, Ledger]
```

## Purchase → Inventory → Journal
```mermaid
flowchart LR
  P1[Create Purchase] --> P2[Submit/Approve]
  P2 --> P3[Receipt (optional)]
  P3 --> P4[Generate Journal: Expense/Inventory + PPN]
  P4 --> P5[Payment: Cash/Transfer/Credit]
  P5 --> P6[Generate Journal: Bank/AP]
  P6 --> P7[Reports: AP Aging, Cash Flow, Ledger]
```

## Asset → Depreciation
```mermaid
flowchart LR
  A1[Asset Purchase] --> A2[Journal: Asset vs Bank/AP]
  A2 --> A3[Monthly Depreciation]
  A3 --> A4[Journal: Depreciation vs Accumulated]
  A4 --> A5[Reports: BS, P&L, Ledger]
```
