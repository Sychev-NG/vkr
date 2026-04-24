package recipe

import (
	"context"
	"errors"
	"strings"

	"vkr/internal/entity"
)

type TxManager interface {
    RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type RecipeSaver interface {
	Add(ctx context.Context, vo entity.UpsertRecipeVO) (*entity.Recipe, error)
	Update(ctx context.Context, id int, vo entity.UpsertRecipeVO) (*entity.Recipe, error)
	Delete(ctx context.Context, id int) (error)
}

type RecipeProvider interface{
	GetById(ctx context.Context, id int) (*entity.Recipe, error)
	GetAll(ctx context.Context) ([]entity.Recipe, error)
}

type ProductProvider interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type RecipeService struct {
	txManager		TxManager
	saver 			RecipeSaver
	provider		RecipeProvider
	productProvider	ProductProvider
}

func New(tx TxManager, rs RecipeSaver, rp RecipeProvider, pp ProductProvider) *RecipeService {
	return &RecipeService{tx, rs, rp, pp}
}

func (ps *RecipeService) Add(ctx context.Context, vo entity.UpsertRecipeVO) (*entity.Recipe, error) {
	vo.Name = strings.TrimSpace(vo.Name)
	if len(vo.Name) == 0 {
		return nil, entity.ErrInvalidRecipeName
	}

	finished, err := ps.productProvider.GetById(ctx, vo.ProductID)
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			return nil, entity.ErrFinishedProductNotFound
		}
		return nil, err
	}

	if entity.ProductType(finished.TypeName) != entity.Finished {
		return nil, entity.ErrInvalidFinishedMaterial
	}
	
	for _, ingredient := range vo.Ingredients {
		raw, err := ps.productProvider.GetById(ctx, ingredient.RawMaterialID)

		if entity.ProductType(raw.TypeName) != entity.Raw {
			return nil, entity.ErrInvalidRawMaterial
		}

		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				return nil, entity.ErrRawProductNotFound
			}
			return nil, err
		}
	}

	var result *entity.Recipe

	err = ps.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		result, err = ps.saver.Add(ctx, vo)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err	
}

func (ps *RecipeService) GetById(ctx context.Context, id int) (*entity.Recipe, error) {
	return ps.provider.GetById(ctx, id)
}

func (ps *RecipeService) GetAll(ctx context.Context) ([]entity.Recipe, error) {
	return ps.provider.GetAll(ctx)
}

func (s *RecipeService) Update(ctx context.Context, id int, vo entity.UpsertRecipeVO) (*entity.Recipe, error) {
    vo.Name = strings.TrimSpace(vo.Name)
    if len(vo.Name) == 0 {
        return nil, entity.ErrInvalidRecipeName
    }
    
	_, err := s.provider.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	finished, err := s.productProvider.GetById(ctx, vo.ProductID)
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			return nil, entity.ErrFinishedProductNotFound
		}
		return nil, err
	}

	if entity.ProductType(finished.TypeName) != entity.Finished {
		return nil, entity.ErrInvalidFinishedMaterial
	}

	for _, ingredient := range vo.Ingredients {
		raw, err := s.productProvider.GetById(ctx, ingredient.RawMaterialID)

		if entity.ProductType(raw.TypeName) != entity.Raw {
			return nil, entity.ErrInvalidRawMaterial
		}

		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				return nil, entity.ErrRawProductNotFound
			}
			return nil, err
		}
	}

    var result *entity.Recipe
    
    err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
        result, err = s.saver.Update(txCtx, id, vo)
        if err != nil {
            return err
        }
        
        return nil
    })
    
    return result, err
}

func (ps *RecipeService) Delete(ctx context.Context, id int) (error) {
	return ps.saver.Delete(ctx, id)
}