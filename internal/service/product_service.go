package service

import (
	"errors"
	"math"
	"time"

	"socialmeli/internal/domain"
	"socialmeli/internal/store"
)

var ErrDateFormat = errors.New("Data inválida. Use dd-MM-aaaa")

type ProductService struct {
	st store.Store
}

func NewProductService(st store.Store) *ProductService { return &ProductService{st: st} }

func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, domain.ErrDateEmpty
	}
	t, err := time.Parse("02-01-2006", dateStr)
	if err != nil {
		return time.Time{}, ErrDateFormat
	}
	return t, nil
}

type PublishPayload struct {
	UserID   int            `json:"user_id"`
	Date     string         `json:"date"`
	Product  domain.Product `json:"product"`
	Category int            `json:"category"`
	Price    float64        `json:"price"`
	HasPromo bool           `json:"has_promo"`
	Discount float64        `json:"discount"`
}

func calcFinalPrice(price, discount float64, hasPromo bool) float64 {
	if !hasPromo {
		return price
	}
	// discount em percentual (ex.: 10 = 10%)
	return math.Round((price*(1-(discount/100)))*100) / 100
}

func (s *ProductService) Publish(payload PublishPayload) (int, error) {
	if err := domain.ValidateID(payload.UserID); err != nil {
		return 0, err
	}
	dt, err := parseDate(payload.Date)
	if err != nil {
		return 0, err
	}

	// validações produto
	if err := domain.ValidateID(payload.Product.ProductID); err != nil {
		return 0, err
	}
	if err := domain.ValidateTextRequired(payload.Product.ProductName, 40, domain.ErrMaxLen40); err != nil {
		return 0, err
	}
	if err := domain.ValidateTextRequired(payload.Product.Type, 15, domain.ErrMaxLen15); err != nil {
		return 0, err
	}
	if err := domain.ValidateTextRequired(payload.Product.Brand, 25, domain.ErrMaxLen25); err != nil {
		return 0, err
	}
	if err := domain.ValidateTextRequired(payload.Product.Color, 15, domain.ErrMaxLen15); err != nil {
		return 0, err
	}
	if err := domain.ValidateNotesOptional(payload.Product.Notes); err != nil {
		return 0, err
	}

	if payload.Category == 0 {
		return 0, domain.ErrCategoryEmpty
	}
	if err := domain.ValidatePrice(payload.Price); err != nil {
		return 0, err
	}

	// valida desconto quando for promocao
	if payload.HasPromo {
		if payload.Discount <= 0 {
			return 0, errors.New("Desconto deve ser maior que zero")
		}
		if payload.Discount > 100 {
			return 0, errors.New("Desconto deve ser menor ou igual a 100")
		}
	} else {
		payload.Discount = 0
	}

	p := domain.Post{
		UserID:     payload.UserID,
		Date:       dt,
		DateStr:    payload.Date,
		Product:    payload.Product,
		Category:   payload.Category,
		Price:      payload.Price,
		HasPromo:   payload.HasPromo,
		Discount:   payload.Discount,
		FinalPrice: calcFinalPrice(payload.Price, payload.Discount, payload.HasPromo),
	}
	return s.st.AddPost(p)
}

func (s *ProductService) FollowedLastTwoWeeks(userID int, order string) ([]domain.Post, error) {
	if err := domain.ValidateID(userID); err != nil {
		return nil, err
	}
	if err := domain.ValidateOrderForPosts(order); err != nil {
		return nil, err
	}

	followed, err := s.st.FollowedBy(userID)
	if err != nil {
		return nil, err
	}

	sellerIDs := make([]int, 0, len(followed))
	for _, u := range followed {
		sellerIDs = append(sellerIDs, u.ID)
	}

	since := time.Now().AddDate(0, 0, -14)
	posts := s.st.PostsFromSellersSince(sellerIDs, since)
	domain.SortPostsByDate(posts, order)
	return posts, nil
}

func (s *ProductService) PromoCount(userID int) (domain.User, int, error) {
	if err := domain.ValidateID(userID); err != nil {
		return domain.User{}, 0, err
	}
	u, ok := s.st.GetUser(userID)
	if !ok {
		return domain.User{}, 0, store.ErrUserNotFound
	}
	posts := s.st.PromoPostsBySeller(userID)
	return u, len(posts), nil
}

func (s *ProductService) PromoList(userID int) (domain.User, []domain.Post, error) {
	if err := domain.ValidateID(userID); err != nil {
		return domain.User{}, nil, err
	}
	u, ok := s.st.GetUser(userID)
	if !ok {
		return domain.User{}, nil, store.ErrUserNotFound
	}
	posts := s.st.PromoPostsBySeller(userID)
	domain.SortPostsByDate(posts, domain.DateDesc)
	return u, posts, nil
}

func (s *ProductService) DeleteMyPost(userID, postID int) error {
	if err := domain.ValidateID(userID); err != nil {
		return err
	}
	if err := domain.ValidateID(postID); err != nil {
		return err
	}
	return s.st.DeletePost(userID, postID)
}
