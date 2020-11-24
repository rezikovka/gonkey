# Gonkey: инструмент автоматизации тестирования

Gonkey протестирует ваши сервисы, используя их API. Он умеет обстреливать сервис заранее заготовленными запросами и проверять ответы. Сценарий теста описывается в YAML-файле.

Возможности:
- работает с REST/JSON API
- проверка API сервиса на соответствие OpenAPI-спеке
- заполнение БД сервиса данными из фикстур (поддерживается PostgreSQL)
- можно подключить к проекту как библиотеку и запускать вместе с юнит-тестами

### Использование консольной утилиты

Для тестирование сервиса, размещенного на удаленном хосте, используйте gonkey как консольную утилиту.

`./gonkey -host <...> -tests <...> [-spec <...>] [-db_dsn <...> -fixtures <...>] [-v]`

- `-spec <...>` путь к файлу или URL со swagger-спецификацией сервиса
- `-host <...>` хост:порт сервиса
- `-tests <...>` файл или директория с тестами
- `-db_dsn <...>` dsn для вашей тестовой базы данных (бд будет очищена перед наполнением!), поддерживается только PostgreSQL
- `-fixtures <...>` директория с вашими фикстурами
- `-v` подробный вывод
- `-debug` отладочный вывод

### Использование gonkey как библиотеки

Чтобы интегрировать функциональные тесты в нативные тесты Go и запускать их вместе, используйте gonkey как библиотеку.

Создайте файл для будущего теста, например, `func_test.go`.

Подключите gonkey как зависимость к вашему проекту в этом файле.

```go
import (
	"github.com/rezikovka/gonkey/runner"
)
```

Создайте функцию с тестом.

```go
func TestFuncCases(t *testing.T) {
    // проинициализирйте базу для загрузки фикстур, если нужно (подробнее - ниже)
    //db := ...

    // создайте экземпляр сервера вашего приложения
    srv := server.NewServer()
    defer srv.Close()

    // запустите выполнение тестов из директории cases с записью в отчет Allure
    runner.RunWithTesting(t, &runner.RunWithTestingParams{
        Server:      srv,
        TestsDir:    "cases",
        DB:          db,
        FixturesDir: "fixtures",
    })
}
```

Теперь тесты можно запускать через `go test`, например, так: `go test ./...`.

### Пример файла с тестами
```yaml
- name: КОГДА запрашивается список заказов ДОЛЖЕН успешно возвращаться
  method: GET
  path: /jsonrpc/v2/order.getBriefList
  query: ?id=550e8400-e29b-41d4-a716-446655440000&jsonrpc=2.0&user_id=00001

  fixtures:
    - order_0001
    - order_0002

  response:
    200: |
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "jsonrpc": "2.0",
        "result": {
          "data": [
            "ORDER0001",
            "ORDER0002"
          ],
          "meta": {
            "items": 0,
            "limit": 50,
            "page": 0,
            "pages": 0
          }
        }
      }

- name: КОГДА запрашивается один заказ ДОЛЖЕН возвращаться пользователь и сумма
  method: POST
  path: /jsonrpc/v2/order.getOrder

  headers:
    Authorization: Bearer HsHG67d38hJKJFdfjj==
    Content-Type: application/json

  cookies:
    sid: ZmEwZDkwYzgwMmQzMGIzOGIxODM3ZmFiOTGJhMzU=
    lid: AAAEAFu/TdhHBg7UAgA=

  comparisonParams:
    ignoreValues: false
    ignoreArraysOrdering: false
    disallowExtraFields: false

  request: |
    {
      "jsonrpc": "2.0",
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "method": "order.getOrder",
      "params": [
        {
          "order_nr": {{ .orderNr }}
        }
      ]
    }

  response:
    200: |
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "jsonrpc": "2.0",
        "result": {
          "user_id": {{ .userId }},
          "amount": {{ .amount }}
        }
      }

responseHeaders:
    200:
      Content-Type: "application/json"
      Cache-Control: "no-store, must-revalidate"
      Set-Cookie: "mycookie=123; Path=/; Domain=mydomain.com", "mycookie=456; Path=/; Domain=.mydomain.com"

  cases:
    - requestArgs:
        orderNr: ORDER0001
      responseArgs:
        200:
          userId: '0001'
          amount: 1000

    - requestArgs:
        orderNr: ORDER0002
      responseArgs:
        200:
          userId: '0001'
          amount: 72000
```

Как видно из примера, вы можете использовать Regexp для проверки тела ответа.
Они могут быть использованы для проверки всего тела (если это просто текст):
```
    response:
        200: "$matchRegexp(^xy+z$)"
```
или для элементов map/array (если это JSON):
```
    response:
        200: |
          {
            "id": "[\w-]+",
            "jsonrpc": "[12].0",
            "result": [
              "data": [
                  "ORDER[0]{3}[0-9]",
                  "ORDER[0]{3}[0-9]"
              ],
            ]
          }
```

### HTTP-запрос

`method` - параметр для передачи типа HTTP запроса, формат передачи указан в примере выше

`path` - параметр для передачи URL-пути, формат передачи указан в примере выше

`headers` - параметр для передачи http-заголовков, формат передачи указан в примере выше.

`cookies` -  параметр для передачи cookie, формат передачи указан в примере выше.

### HTTP-ответ

`response` - тело ответа HTTP для указанных кодов состояния HTTP.

`responseHeaders` - все заголовки ответа HTTP для указанных кодов состояния HTTP.

### Переменные

В описании теста можно использовать переменные, они поддерживаются в следующих полях:

- method
- path
- query
- headers
- request
- response

Пример использования:

```yaml
- method: "{{ $method }}"
  path: "/some/path/{{ $pathPart }}"
  query: "{{ $query }}"
  headers:
    header1: "{{ $header }}"
  request: '{"reqParam": "{{ $reqParam }}"}'
  response:
    200: "{{ $resp }}"
```

Присваивать значения переменным можно следующими способами:

- в описании самого теста
- из результатов предыдущего запроса
- в переменных окружения или в env-файле

Приоритеты источников соответствуют порядку перечисления.

#### Подробнее про способы присваивания:

##### В описании самого теста

Пример: 

```yaml
- method: "{{ $method }}"
  path: "/some/path/{{ $pathPart }}"
  variables:
    reqParam: "reqParam_value"
    method: "POST"
    pathPart: "part_of_path"
    query: "query_val"
    header: "header_val"
    resp: "resp_val"
  query: "{{ $query }}"
  headers:
    header1: "{{ $header }}"
  request: '{"reqParam": "{{ $reqParam }}"}'
  response:
    200: "{{ $resp }}"
```

##### Из результатов предыдущего запроса

Пример:

```yaml
# если в ответе plain text
- name: "get_last_post_id"
  ...
  variables_to_set:
          200: "id"

# если в ответе JSON
- name: "get_last_post_info"
  variables_to_set:
          200:
            id: "id"
            title: "title"
            authorId: "author_info.id"
```

Обратите внимание - если нужно использовать значение вложенного поля, можно указать путь до него:
> "author_info.id"

Глубина вложенности может быть любая.

##### В переменных окружения или в env-файле

Gonkey автоматически проверяет наличие указанной переменной среди переменных окружения (в таком же регистре) и берет значение оттуда, в случае наличия.

Если указан env-файл, то описанные в нем переменные добавятся/заменят соответствующие перемнные окружения.
env-файл указывается с помощью параметра env-file

Пример env-файла (стандартный синтаксис):
```.env
jwt=some_jwt_value
secret=my_secret
password=private_password
```

env-файл, например, удобно использовать, когда нужно вынести из теста приватную информацию (пароли, ключи и т.п.)


### Загрузка файлов

В тестовом запросе можно загружать файлы. Для этого нужно указать тип запроса - POST и заголовок:

> Content-Type: multipart/form-data

Пример:

```yaml
 - name: "upload-files"
   method: POST
   form:
       files:
         file1: "testdata/upload-files/file1.txt"
         file2: "testdata/upload-files/file2.log"
   headers:
     Content-Type: multipart/form-data # case-sensitive, can be omitted
   response:
     200: |
       {
         "status": "OK"
       }
```

### Фикстуры

Чтобы наполнить базу перед тестом, используются файлы с фикстурами.

Пример файла:

```yaml
# fixtures/comments.yml
inherits:
  - another_fixture
  - yet_another_fixture

tables:
  posts:
    - $name: janes_post
      title: New post
      text: Post text
      author: Jane Dow
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

    - $name: apples_post
      title: Morning digest
      text: Text
      author: Apple Seed
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

  comments:
    - post_id: $janes_post.id
      content: A comment...
      author_name: John Doe
      author_email: john@doe.com
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

    - post_id: $apples_post.id
      content: Another comment...
      author_name: John Doe
      author_email: john@doe.com
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

  another_table:
      ...
  ...
```

Записи в фикстурах можно наследовать одну от другой, использовать шаблоны, а так же ссылаться из одной записи на другую.

#### Шаблоны записей

Обычно, чтобы вставить строку данных в базу, вам нужно перечислить все поля, для которых в базе не предусмотрено значение по умолчанию. Довольно часто, многие из этих полей не важны для теста и их значения повторяются от одной фикстуры к другой, создавая ненужный визуальный мусор и усложняя их поддержку.

С помощью шаблонов вы можете наследовать поля из шаблонной записи, каждый раз переопределяя только те поля, которые действительно важны для теста.

Пример определения шаблона:
```yaml
templates:
  dummy_client:
    name: Dummy Client Name
    age: 35
    ip: 127.0.0.1
    is_deleted: false

  dummy_deleted_client:
    $extend: dummy_client
    is_deleted: true

tables:
  ...
```

Пример использования шаблона в фикстуре:
```yaml
templates:
   ...
tables:
   clients:
      - $extend: dummy_client
      - $extend: dummy_client
        name: Josh
      - $extend: dummy_deleted_client
        name: Jane
```

Как вы могли заметить, шаблоны тоже можно наследовать друг от друга с помощью ключевого слова `$extend`, но только если к моменту определения зависимого шаблона базовый шаблон уже определен (в этом же файле или любом, подключенном через `inherits`).

#### Наследование записей

Записи, как и шаблоны, можно наследовать с помощью `$extend`.

Для того, чтобы унаследовать запись, сначала нужно присвоить строке имя с помощью `$name`:

```yaml
# fixtures/post.yaml
tables:
  posts:
    - $name: regular_post
      title: Post title
      text: Some text
```

Имена, присваиваемые строкам, должны быть уникальны среди всех загружаемых файлов с фикстурами, а также не пересекаться с именами шаблонов.

В другом файле фикстур нужно объявить, что определенная строка наследует ранее объявленную запись с помощью `$extend`, так же, как и в случае с шаблоном:

```yaml
# fixtures/deleted_post.yaml
inherits:
  - post
tables:
  posts:
    - $extend: regular_post
      is_deleted: true
```

Не забудьте объявить зависимость между файлами в `inherits`, чтобы один файл всегда загружался вместе с другим.

Обратите внимание, что наследование строк работает только между разными файлами фикстур. В пределах одного файла объявить наследование невозможно.

#### Связывание записей

Несмотря на то, что файлы фикстур позволяют вам задавать значение для автоинкрементных колонок (обычно `id`), не рекомендуется этого делать. Контролировать, чтобы в разных файлах использовались правильные значения для `id`, и чтобы они нигде не пересеклись, очень сложно. Чтобы база сама присвоила значение автоинкрементному полю, достаточно просто не указывать его значение явно.

Однако, если не указывать значение для `id`, то как связать несколько сущностей, которые должны ссылаться друг на друга по идентификаторам? Фикстуры позволяют ссылаться на значение ранее вставленных в базу строк по их имени, используя нотацию `$refName.fieldName`.

Объявим именованную запись:

```yaml
# fixtures/post.yaml
tables:
  posts:
    - $name: regular_post
      title: Post title
      text: Some text
```

Теперь, чтобы связать таблицы `posts` и `comments`, обратимся к записи по имени (`$regular_post`) и укажем поле, из которого следует взять значение (`id`):

```yaml
# fixtures/comment.yaml
tables:
  comments:
    - post_id: $regular_post.id
      content: A comment...
      author_name: John Doe
```

Ссылаться можно только на поля ранее вставленной в базу записи, на поля шаблона ссылаться нельзя, при попытке это сделать, вы получите ошибку `undefined reference`.

Обратите внимание на ограничение: нельзя ссылаться на записи в пределах одной таблицы одного файла.

#### Выражения

Если в базу нужно записать не статичное значение, а результат исполнения выражения, то можно воспользоваться конструкцией `$eval()`. Все, что будет задано внутри скобок, будет вставлено в базу в сыром, неэкранированном виде. Таким образом, внутри `$eval()` можно написать все то, что вы могли бы написать в самом запросе.

Например, такая конструкция вставит текущую дату и время в качестве значения поля:

```yaml
tables:
  comments:
    - created_at: $eval(NOW())
```

### Запрос в Базу данных

После выполнения http запросов можно выполнить SQL запрос в БД для проверки изменений данных. 
Допускается что ответ может содержать несколько записей. Далее эти данные сравниваются с ожидаемым списком записей.

#### Описание запроса

Под запросом подразумевается SELECT, который вернет любое количество строк.

- `dbQuery` - строка, содержит SQL запрос

Пример:
```yaml
  ...
  dbQuery:
    SELECT code, purchase_date, partner_id FROM mark_paid_schedule AS m WHERE m.code = 'GIFT100000-000002'
  ...
```

#### Описание ответа на запрос в Базу данных

Под ответом подразумевается список json объектов которые должен вернуть запрос в БД.

- `dbResponse` - строка, содержит список json объектов

Пример:
```yaml
  ...
  dbResponse:
    - '{"code":"GIFT100000-000002","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
    - '{"code":"GIFT100000-000003","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
```

```yaml
  ...
  dbResponse:
    # пустой список
```
#### Параметризация при запросах в Базу данных

Как и в случае с телом http-запроса, мы можем использовать параметризированные запросы.

Пример:
```yaml
  ...
    dbQuery: >
      SELECT code, partner_id FROM mark_paid_schedule AS m WHERE DATE(m.purchase_date) BETWEEN '{{ .fromDate }}' AND '{{ .toDate }}'

    dbResponse:
      - '{"code":"{{ .cert1 }}","partner_id":1}'
      - '{"code":"{{ .cert2 }}","partner_id":1}'
  ...
    cases:
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseArgs:
        cert1: "GIFT100000-000002"
        cert2: "GIFT100000-000003"
```

В случае, когда в разных тестах ответ содержит разное количество записей, вы можете переопределить ответ целиком для конкретного теста 
продолжая использовать шаблон с параметрами в остальных

Пример:
```yaml
  ...
    dbQuery: >
      SELECT code, partner_id FROM mark_paid_schedule AS m WHERE DATE(m.purchase_date) BETWEEN '{{ .fromDate }}' AND '{{ .toDate }}'

    dbResponse:
      - '{"code":"{{ .cert1 }}","partner_id":1}'
  ...
    cases:
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseArgs:
        cert1: "GIFT100000-000002"
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseFull:
        - '{"code":"GIFT100000-000002","partner_id":1}'
        - '{"code":"GIFT100000-000003","partner_id":1}'
```

