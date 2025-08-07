// MongoDB Collections Design for Multi-tenant Architecture
// Database-per-tenant approach
// Each tenant gets their own MongoDB database

// Collection schemas and indexes for tenant databases

// 1. Analytics Collection
// Database: tenant_{tenant_slug}_analytics
db.createCollection("page_views", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["tenant_id", "user_id", "page", "timestamp"],
      properties: {
        tenant_id: {
          bsonType: "string",
          description: "Tenant ID is required"
        },
        user_id: {
          bsonType: "string",
          description: "User ID is required"
        },
        page: {
          bsonType: "string",
          description: "Page URL is required"
        },
        timestamp: {
          bsonType: "date",
          description: "Timestamp is required"
        },
        session_id: {
          bsonType: "string"
        },
        ip_address: {
          bsonType: "string"
        },
        user_agent: {
          bsonType: "string"
        },
        referrer: {
          bsonType: "string"
        },
        duration: {
          bsonType: "int",
          description: "Time spent on page in seconds"
        },
        metadata: {
          bsonType: "object"
        }
      }
    }
  }
});

// Create indexes for page_views
db.page_views.createIndex({ "tenant_id": 1 });
db.page_views.createIndex({ "user_id": 1 });
db.page_views.createIndex({ "timestamp": -1 });
db.page_views.createIndex({ "page": 1, "timestamp": -1 });
db.page_views.createIndex({ "session_id": 1 });

// 2. User Events Collection
db.createCollection("user_events", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["tenant_id", "user_id", "event_type", "timestamp"],
      properties: {
        tenant_id: {
          bsonType: "string"
        },
        user_id: {
          bsonType: "string"
        },
        event_type: {
          bsonType: "string",
          enum: ["login", "logout", "purchase", "view_product", "add_to_cart", "remove_from_cart", "search", "custom"]
        },
        timestamp: {
          bsonType: "date"
        },
        session_id: {
          bsonType: "string"
        },
        properties: {
          bsonType: "object"
        },
        metadata: {
          bsonType: "object"
        }
      }
    }
  }
});

// Create indexes for user_events
db.user_events.createIndex({ "tenant_id": 1 });
db.user_events.createIndex({ "user_id": 1, "timestamp": -1 });
db.user_events.createIndex({ "event_type": 1, "timestamp": -1 });
db.user_events.createIndex({ "session_id": 1 });

// 3. Notifications Collection
db.createCollection("notifications", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["tenant_id", "user_id", "type", "title", "created_at"],
      properties: {
        tenant_id: {
          bsonType: "string"
        },
        user_id: {
          bsonType: "string"
        },
        type: {
          bsonType: "string",
          enum: ["info", "warning", "error", "success", "promotion", "system"]
        },
        title: {
          bsonType: "string"
        },
        message: {
          bsonType: "string"
        },
        is_read: {
          bsonType: "bool"
        },
        read_at: {
          bsonType: "date"
        },
        action_url: {
          bsonType: "string"
        },
        priority: {
          bsonType: "string",
          enum: ["low", "medium", "high", "urgent"]
        },
        created_at: {
          bsonType: "date"
        },
        expires_at: {
          bsonType: "date"
        },
        metadata: {
          bsonType: "object"
        }
      }
    }
  }
});

// Create indexes for notifications
db.notifications.createIndex({ "tenant_id": 1 });
db.notifications.createIndex({ "user_id": 1, "created_at": -1 });
db.notifications.createIndex({ "is_read": 1, "created_at": -1 });
db.notifications.createIndex({ "type": 1, "created_at": -1 });
db.notifications.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 });

// 4. Chat Messages Collection
db.createCollection("chat_messages", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["tenant_id", "room_id", "sender_id", "message", "timestamp"],
      properties: {
        tenant_id: {
          bsonType: "string"
        },
        room_id: {
          bsonType: "string"
        },
        sender_id: {
          bsonType: "string"
        },
        message: {
          bsonType: "string"
        },
        message_type: {
          bsonType: "string",
          enum: ["text", "image", "file", "system"]
        },
        timestamp: {
          bsonType: "date"
        },
        edited_at: {
          bsonType: "date"
        },
        is_deleted: {
          bsonType: "bool"
        },
        attachments: {
          bsonType: "array",
          items: {
            bsonType: "object",
            properties: {
              file_id: { bsonType: "string" },
              filename: { bsonType: "string" },
              file_size: { bsonType: "int" },
              mime_type: { bsonType: "string" }
            }
          }
        },
        reactions: {
          bsonType: "array",
          items: {
            bsonType: "object",
            properties: {
              emoji: { bsonType: "string" },
              user_id: { bsonType: "string" },
              timestamp: { bsonType: "date" }
            }
          }
        },
        metadata: {
          bsonType: "object"
        }
      }
    }
  }
});

// Create indexes for chat_messages
db.chat_messages.createIndex({ "tenant_id": 1 });
db.chat_messages.createIndex({ "room_id": 1, "timestamp": -1 });
db.chat_messages.createIndex({ "sender_id": 1, "timestamp": -1 });

// 5. Activity Logs Collection
db.createCollection("activity_logs", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["tenant_id", "user_id", "action", "resource_type", "timestamp"],
      properties: {
        tenant_id: {
          bsonType: "string"
        },
        user_id: {
          bsonType: "string"
        },
        action: {
          bsonType: "string",
          enum: ["create", "read", "update", "delete", "login", "logout", "export", "import"]
        },
        resource_type: {
          bsonType: "string",
          enum: ["user", "product", "order", "category", "file", "tenant", "system"]
        },
        resource_id: {
          bsonType: "string"
        },
        timestamp: {
          bsonType: "date"
        },
        ip_address: {
          bsonType: "string"
        },
        user_agent: {
          bsonType: "string"
        },
        changes: {
          bsonType: "object"
        },
        metadata: {
          bsonType: "object"
        }
      }
    }
  }
});

// Create indexes for activity_logs
db.activity_logs.createIndex({ "tenant_id": 1 });
db.activity_logs.createIndex({ "user_id": 1, "timestamp": -1 });
db.activity_logs.createIndex({ "action": 1, "timestamp": -1 });
db.activity_logs.createIndex({ "resource_type": 1, "resource_id": 1 });
db.activity_logs.createIndex({ "timestamp": -1 });

// 6. Cache Collection (for application-level caching)
db.createCollection("cache", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["key", "value", "created_at"],
      properties: {
        key: {
          bsonType: "string"
        },
        value: {
          bsonType: "object"
        },
        created_at: {
          bsonType: "date"
        },
        expires_at: {
          bsonType: "date"
        },
        tags: {
          bsonType: "array",
          items: {
            bsonType: "string"
          }
        }
      }
    }
  }
});

// Create indexes for cache
db.cache.createIndex({ "key": 1 }, { unique: true });
db.cache.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 });
db.cache.createIndex({ "tags": 1 });

// 7. Search Index Collection (for full-text search)
db.createCollection("search_index", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["tenant_id", "resource_type", "resource_id", "content"],
      properties: {
        tenant_id: {
          bsonType: "string"
        },
        resource_type: {
          bsonType: "string",
          enum: ["product", "user", "order", "category", "file"]
        },
        resource_id: {
          bsonType: "string"
        },
        content: {
          bsonType: "string"
        },
        tags: {
          bsonType: "array",
          items: {
            bsonType: "string"
          }
        },
        created_at: {
          bsonType: "date"
        },
        updated_at: {
          bsonType: "date"
        }
      }
    }
  }
});

// Create indexes for search_index
db.search_index.createIndex({ "tenant_id": 1 });
db.search_index.createIndex({ "resource_type": 1, "resource_id": 1 }, { unique: true });
db.search_index.createIndex({ "content": "text", "tags": "text" });

// 8. Reports Collection (for storing generated reports)
db.createCollection("reports", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["tenant_id", "user_id", "report_type", "created_at"],
      properties: {
        tenant_id: {
          bsonType: "string"
        },
        user_id: {
          bsonType: "string"
        },
        report_type: {
          bsonType: "string",
          enum: ["sales", "inventory", "users", "analytics", "financial"]
        },
        title: {
          bsonType: "string"
        },
        parameters: {
          bsonType: "object"
        },
        data: {
          bsonType: "object"
        },
        file_url: {
          bsonType: "string"
        },
        status: {
          bsonType: "string",
          enum: ["pending", "processing", "completed", "failed"]
        },
        created_at: {
          bsonType: "date"
        },
        completed_at: {
          bsonType: "date"
        },
        expires_at: {
          bsonType: "date"
        }
      }
    }
  }
});

// Create indexes for reports
db.reports.createIndex({ "tenant_id": 1 });
db.reports.createIndex({ "user_id": 1, "created_at": -1 });
db.reports.createIndex({ "report_type": 1, "created_at": -1 });
db.reports.createIndex({ "status": 1 });
db.reports.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 });

// Print initialization complete message
print("MongoDB collections and indexes created successfully for tenant database");
