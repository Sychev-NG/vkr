package assembly

import (
	"context"
	"errors"
	"log"

	"vkr/internal/entity"
	"vkr/internal/repository/postgres"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type ProductProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type AssemblyService struct {
	txManager       TxManager
	f               *postgres.RepositoryFactory
	productProvider ProductProvider
}

func New(
	txManager TxManager,
	f *postgres.RepositoryFactory,
	pp ProductProvider,
) *AssemblyService {
	return &AssemblyService{txManager, f, pp}
}

func (s *AssemblyService) Add(ctx context.Context, vo entity.UpsertAssemblyVO) error {
	for _, item := range vo.Components {
		_, err := s.productProvider.GetByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				return entity.ErrProductNotFound
			}
			log.Printf("AssemblyService::Add productProvider.GetByID Error - %v", err)
			return err
		}
	}

	_, err := s.productProvider.GetByID(ctx, vo.OutputProductID)
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			return entity.ErrProductNotFound
		}
		log.Printf("AssemblyService::Add output product validation Error - %v", err)
		return err
	}

	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		aRepo := s.f.NewAssemblyRepository(txCtx)
		assemblyID, err := aRepo.Create(txCtx, vo)
		if err != nil {
			log.Printf("AssemblyService::Add Create Error - %v", err)
			return err
		}

		for _, component := range vo.Components {
			err := aRepo.AddComponent(txCtx, assemblyID, component.ProductID, component.Quantity)
			if err != nil {
				log.Printf("AssemblyService::Add AddComponent Error - %v", err)
				return err
			}
		}

		return nil
	})

	return err
}

func (s *AssemblyService) GetByID(ctx context.Context, id int) (*entity.Assembly, error) {
	aRepo := s.f.NewAssemblyRepository(ctx)
	assembly, err := aRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrAssemblyNotFound) {
			return nil, entity.ErrAssemblyNotFound
		}
		log.Printf("AssemblyService::GetByID Error - %v", err)
		return nil, err
	}
	return assembly, nil
}

func (s *AssemblyService) GetAll(ctx context.Context) ([]entity.Assembly, error) {
	aRepo := s.f.NewAssemblyRepository(ctx)
	assemblies, err := aRepo.GetAll(ctx)
	if err != nil {
		log.Printf("AssemblyService::GetAll Error - %v", err)
		return nil, err
	}
	return assemblies, nil
}

func (s *AssemblyService) Update(ctx context.Context, id int, vo entity.UpsertAssemblyVO) error {
	aRepo := s.f.NewAssemblyRepository(ctx)
	_, err := aRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrAssemblyNotFound) {
			return entity.ErrAssemblyNotFound
		}
		log.Printf("AssemblyService::Update GetByID Error - %v", err)
		return err
	}

	for _, item := range vo.Components {
		_, err := s.productProvider.GetByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				return entity.ErrProductNotFound
			}
			log.Printf("AssemblyService::Update productProvider.GetByID Error - %v", err)
			return err
		}
	}

	_, err = s.productProvider.GetByID(ctx, vo.OutputProductID)
	if err != nil {
		if errors.Is(err, entity.ErrProductNotFound) {
			return entity.ErrProductNotFound
		}
		log.Printf("AssemblyService::Update output product validation Error - %v", err)
		return err
	}

	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		aRepo := s.f.NewAssemblyRepository(txCtx)

		err := aRepo.Update(txCtx, id, vo)
		if err != nil {
			log.Printf("AssemblyService::Update Update Error - %v", err)
			return err
		}

		components := make([]struct {
			ProductID int
			Quantity  float64
		}, len(vo.Components))

		for i, comp := range vo.Components {
			components[i].ProductID = comp.ProductID
			components[i].Quantity = comp.Quantity
		}

		err = aRepo.ReplaceComponents(txCtx, id, components)
		if err != nil {
			log.Printf("AssemblyService::Update ReplaceComponents Error - %v", err)
			return err
		}

		return nil
	})

	return err
}

func (s *AssemblyService) Delete(ctx context.Context, id int) error {
	aRepo := s.f.NewAssemblyRepository(ctx)
	err := aRepo.Delete(ctx, id)
	if err != nil {
		log.Printf("AssemblyService::Delete Delete Error - %v", err)
		return err
	}

	return nil
}