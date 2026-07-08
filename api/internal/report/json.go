package report

import (
	"encoding/json"
	"time"

	"github.com/re-conta/reconta/api/internal/transaction"
)

// ToBackupPayload achata as transações (que referenciam categoria/conta por
// id) para um formato que referencia por nome, portável entre bancos de dados.
// accountNames mapeia o id de cada conta do usuário para o seu nome.
func ToBackupPayload(scope Scope, txs []transaction.Transaction, totals transaction.Totals, accountNames map[int64]string) BackupPayload {
	categorySeen := map[string]bool{}
	categories := []CategoryRef{}
	accountSeen := map[string]bool{}
	accounts := []AccountRef{}
	tagSeen := map[string]bool{}
	tags := []TagRef{}

	records := make([]TxRecord, 0, len(txs))
	for _, tx := range txs {
		rec := TxRecord{
			Date:           tx.Date,
			Description:    tx.Description,
			Amount:         tx.Amount,
			Type:           tx.Type,
			Notes:          tx.Notes,
			PixBeneficiary: tx.PixBeneficiary,
			Bank:           tx.Bank,
		}
		if tx.AccountID != nil {
			if name, ok := accountNames[*tx.AccountID]; ok {
				rec.AccountName = &name
				if !accountSeen[name] {
					accountSeen[name] = true
					accounts = append(accounts, AccountRef{Name: name})
				}
			}
		}
		if tx.CategoryName != nil {
			rec.CategoryName = tx.CategoryName
			if !categorySeen[*tx.CategoryName] {
				categorySeen[*tx.CategoryName] = true
				color := ""
				if tx.CategoryColor != nil {
					color = *tx.CategoryColor
				}
				categories = append(categories, CategoryRef{
					Name: *tx.CategoryName, Color: color, Type: tx.Type,
				})
			}
		}
		for _, t := range tx.Tags {
			rec.Tags = append(rec.Tags, t.Name)
			if !tagSeen[t.Name] {
				tagSeen[t.Name] = true
				tags = append(tags, TagRef{Name: t.Name, Color: t.Color})
			}
		}
		records = append(records, rec)
	}

	return BackupPayload{
		Version:      backupVersion,
		ExportedAt:   time.Now().UTC(),
		PeriodLabel:  scope.Label,
		Categories:   categories,
		Accounts:     accounts,
		Tags:         tags,
		Transactions: records,
		Totals:       totals,
	}
}

// BuildJSON serializa o payload de backup como JSON indentado.
func BuildJSON(payload BackupPayload) ([]byte, error) {
	return json.MarshalIndent(payload, "", "  ")
}
