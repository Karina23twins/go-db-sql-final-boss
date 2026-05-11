package main

import (
	"database/sql"
	// "io/ioutil"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "проверяем что нет ошибки")

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	service := NewParcelService(store)

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	test_p, err := service.Register(parcel.Client, parcel.Address)

	require.NoError(t, err, "проверяем что нет ошибки")
	require.NotEqual(t, 0, test_p.Number)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	test_p, err = service.store.Get(test_p.Number)
	assert.NoError(t, err, "проверяем что нет ошибки")
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	assert.Equal(t, parcel.Address, test_p.Address)
	assert.Equal(t, parcel.Client, test_p.Client)
	assert.Equal(t, parcel.CreatedAt, test_p.CreatedAt)
	assert.NotEqual(t, 0, test_p.Number)
	assert.Equal(t, parcel.Status, test_p.Status)

	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	err = service.store.Delete(test_p.Number)
	require.NoError(t, err, "проверяем что нет ошибки")
	// проверьте, что посылку больше нельзя получить из БД
	_, err = service.store.Get(test_p.Number)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	store := NewParcelStore(db)
	parcel := getTestParcel()
	service := NewParcelService(store)

	test_p, err := service.Register(parcel.Client, parcel.Address)

	require.NoError(t, err, "проверяем что нет ошибки")
	require.NotEqual(t, 0, test_p.Number)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = service.ChangeAddress(test_p.Number, newAddress)
	require.NoError(t, err, "проверяем что нет ошибки")

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	test_p, err = service.store.Get(test_p.Number)
	assert.Equal(t, newAddress, test_p.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	// настройте подключение к БД

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	service := NewParcelService(store)

	test_p, err := service.Register(parcel.Client, parcel.Address)

	require.NoError(t, err, "проверяем что нет ошибки")
	require.NotEqual(t, 0, test_p.Number)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = service.NextStatus(test_p.Number)
	require.NoError(t, err, "проверяем что нет ошибки")

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	test_p, err = service.store.Get(test_p.Number)
	assert.Equal(t, ParcelStatusSent, test_p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// настройте подключение к БД

	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := service.store.Add(parcels[i])
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err, "проверяем что нет ошибки")
		require.NotEqual(t, 0, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := service.store.GetByClient(client)
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	assert.NoError(t, err, "проверяем что нет ошибки")
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Len(t, storedParcels, len(parcelMap))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		test_parcel := parcelMap[parcel.Number]
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		assert.Equal(t, parcel, test_parcel)
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, parcel.Address, test_parcel.Address)
		assert.Equal(t, parcel.Client, test_parcel.Client)
		assert.Equal(t, parcel.CreatedAt, test_parcel.CreatedAt)
		assert.Equal(t, parcel.Number, test_parcel.Number)
		assert.Equal(t, parcel.Status, test_parcel.Status)
	}
}
