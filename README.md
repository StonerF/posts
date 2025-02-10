# Ozon_Test_Task 

Файл `.env` уже находится в корневой директории проекта. 



Для выбора типа хранилища необходимо поменять значение переменной `STORAGE`.
   Возможные параметры:

   - IN_MEMORY - хранение данных in_memory
   - PostgreSQL - хранение данных в бд
   

 **Docker и Docker-compose**
   
   Для запуска сервиса небходимо включить docker и запусить контейнеры командой:
   ```bash
   docker-compose up --build
   ```
   После запуска подключиться к GraphQL Playground можно будет подключиться по адресу:

   ```bash
   http://localhost:8888/
   ```

**Примеры запросов**

- Создание поста:
  ```bash
   mutation {
      createPost(authorId:"1", title:"title", content: "content", allowComments: true) {
         id
         title
         content
         allowComments
      }
   }
   ```

- Создание комментария:
  ```bash
   mutation {
      createComment(authorId:"2", postId:"id поста на котором оставляем комментарий", content: "i am content") {
         id
      }
   }
   ```

- Получение постов:
  ```bash
   query GetPosts {
      posts(first: 10) {
         edges {
            cursor
            node {
               id
               title
               content
               allowComments
            }
         }
         pageInfo {
            hasNextPage
            endCursor
         }
      }
  }
  ```

- Создание ответа:
   ```bash
   mutation {
      createReply(authorId:"12", postId:"id поста, на комментарий которого мы хотим добавить ответ", content: "i am content") {
         id
      }
   }
   ```
  
- Получение комментариев под постом:
   ```bash
   query getComment {
      comments(postId:"1") {
         edges{
            node{
               id
               postId
               parentId
               replies(first:10) {
                  edges {
                     node {
                        id
                        postId
                        parentId
                        replies {
                           edges {
                              node {
                                 id
                                 parentId
                              }
                           }
                        }
                     }
                  }
               }
            }
         }
         pageInfo {
            endCursor
         }
      }
   }
   ```