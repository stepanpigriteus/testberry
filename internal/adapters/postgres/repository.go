package postgres

import (
	"context"
	"database/sql"
	"fmt"
	order_entity "testberry/internal/domain/order"
	"testberry/internal/ports"
	"time"
)

type Repository struct {
	db     *sql.DB
	logger ports.Logger
}

func NewRepository(db *sql.DB, logger ports.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

func (r *Repository) SaveOrder(ctx context.Context, order order_entity.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to start transaction", err)
		return err
	}
	committed := false
	defer func() {
		if err != nil && !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.logger.Error("Failed to rollback transaction", rollbackErr)
			}
		}
	}()

	var deliveryID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO delivery (name, phone, zip, city, address, region, email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	).Scan(&deliveryID)
	if err != nil {
		r.logger.Error("Repo: Failed to insert delivery", err)
		return err
	}

	var paymentID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	).Scan(&paymentID)
	if err != nil {
		r.logger.Error("Repo: Failed to insert payment", err)
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO orders (order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		deliveryID,
		paymentID,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		r.logger.Error("Repo: Failed to insert order", err)
		return err
	}

	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO item (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, order_uid)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
			order.OrderUID,
		)
		if err != nil {
			r.logger.Error("Repo: Failed to insert item", err)
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		r.logger.Error("Repo: Failed to commit transaction", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	r.logger.Info("Repo: Order saved successfully", "order_uid", order.OrderUID)
	return nil
}

func (r *Repository) GetOrderByID(ctx context.Context, orderUID string) (order_entity.Order, error) {
	var order order_entity.Order
	query := `SELECT 
		o.order_uid, o.track_number, o.entry,
		o.locale, o.internal_signature, o.customer_id,
		o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
		d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
		p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
		p.bank, p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders o
	JOIN delivery d ON o.delivery_id = d.id
	JOIN payment p ON o.payment_id = p.id
	WHERE o.order_uid = $1
	`
	err := r.db.QueryRowContext(ctx, query, orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDt,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return order, err
	}

	itemsQuery := `
	SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
	FROM item
	WHERE order_uid = $1
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, orderUID)
	if err != nil {
		return order, err
	}
	defer rows.Close()
	for rows.Next() {
		var item order_entity.Item
		if err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		); err != nil {
			return order, err
		}
		order.Items = append(order.Items, item)
	}
	return order, nil
}

func (r *Repository) RestoreCache(ctx context.Context) ([]order_entity.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
		       o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
		       d.id, d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
		       p.id, p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank,
		       p.delivery_cost, p.goods_total, p.custom_fee
		FROM orders o
		JOIN delivery d ON o.delivery_id = d.id
		JOIN payment p ON o.payment_id = p.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []order_entity.Order

	for rows.Next() {
		var o order_entity.Order
		var d order_entity.Delivery
		var p order_entity.Payment
		var deliveryID, paymentID int
		var dateCreated time.Time

		err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
			&o.CustomerID, &o.DeliveryService, &o.Shardkey, &o.SmID, &dateCreated, &o.OofShard,
			&deliveryID, &d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
			&paymentID, &p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDt, &p.Bank,
			&p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
		)
		if err != nil {
			return nil, err
		}

		o.Delivery = d
		o.Payment = p
		o.DateCreated = dateCreated

		itemRows, err := r.db.QueryContext(ctx, `
			SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
			FROM item WHERE order_uid = $1
		`, o.OrderUID)
		if err != nil {
			return nil, err
		}

		var items []order_entity.Item
		for itemRows.Next() {
			var it order_entity.Item
			err := itemRows.Scan(&it.ChrtID, &it.TrackNumber, &it.Price, &it.Rid, &it.Name, &it.Sale,
				&it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status)
			if err != nil {
				itemRows.Close()
				return nil, err
			}
			items = append(items, it)
		}
		itemRows.Close()
		o.Items = items

		orders = append(orders, o)
	}

	return orders, nil
}
