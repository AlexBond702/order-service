package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AlexBond702/order-service/internal/app/entity"
	"github.com/AlexBond702/order-service/internal/pkg/binding"
	"github.com/AlexBond702/order-service/internal/pkg/http/respondent"
)

func makeErrorMiddleware() gin.HandlerFunc {
	const (
		_40001 = "Bad request"
		_40002 = "Заказ с таким номером уже существует"
		_40003 = "Товар не найден"
		_40004 = "Некорректный статус заказа"
		_40005 = "Цена товара не совпадает"
		_40006 = "Цена товара изменилась"
		_40401 = "Not found"
		_50001 = "Internal sever error"
		_50301 = "Catalog service unavailable"
	)
	const (
		_40401D = "Entity was deleted or never exist"
		_50001D = "Try again later"
		_50301D = "Catalog service is temporarily unavailable"
	)
	makeFallbackManifest := func(status, errorCode int, message, detail string) respondent.ManifestExtractor {
		genericManifest := &respondent.Manifest{
			Status:      status,
			ErrorCode:   errorCode,
			Error:       message,
			ErrorDetail: detail,
		}
		return func(_ error) *respondent.Manifest {
			return genericManifest
		}
	}
	expander := respondent.NewSimpleExpander().ExtractorFor(
		binding.ErrValidatorFailed, binding.NewRespondentManifestExtractor(http.StatusBadRequest, 40001, _40001)).
		WithoutDetail(entity.ErrOrderDuplicate, http.StatusBadRequest, 40002, _40002).
		WithoutDetail(entity.ErrProductNotFound, http.StatusBadRequest, 40003, _40003).
		WithoutDetail(entity.ErrProductNotFoundInCatalog, http.StatusNotFound, 40003, _40003).
		WithoutDetail(entity.ErrProductPriceMismatch, http.StatusBadRequest, 40005, _40005).
		WithoutDetail(entity.ErrProductPriceChanged, http.StatusBadRequest, 40006, _40006).
		WithoutDetail(entity.ErrInvalidOrderStatus, http.StatusBadRequest, 40004, _40004).
		WithDetail(entity.ErrNotFound, http.StatusNotFound, 40401, _40401, _40401D).
		WithDetail(entity.ErrCatalogServiceUnavailable, http.StatusServiceUnavailable, 50301, _50301, _50301D).
		FallbackExtractor(makeFallbackManifest(http.StatusInternalServerError, 50001, _50001, _50001D))

	applicator := respondent.NewSimpleApplicator()

	return respondent.NewMiddleware(expander, applicator)
}
