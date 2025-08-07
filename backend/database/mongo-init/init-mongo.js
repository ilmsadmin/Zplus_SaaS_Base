// MongoDB initialization script for multi-tenant setup
// This script runs when MongoDB container starts for the first time

// Switch to admin database for initial setup
db = db.getSiblingDB('admin');

// Create application user
db.createUser({
  user: 'zplus_app',
  pwd: 'zplus_app_password_123',
  roles: [
    { role: 'readWrite', db: 'zplus_saas_base' },
    { role: 'dbAdmin', db: 'zplus_saas_base' }
  ]
});

// Switch to application database
db = db.getSiblingDB('zplus_saas_base');

// Create collections with initial structure

// Tenant-specific collections (database-per-tenant approach)
// These are examples - actual tenant databases will be created dynamically

// Create system-wide collections
db.createCollection('system_settings', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['key', 'value'],
      properties: {
        key: {
          bsonType: 'string',
          description: 'Setting key must be a string and is required'
        },
        value: {
          description: 'Setting value is required'
        },
        tenant_id: {
          bsonType: 'string',
          description: 'Tenant ID for tenant-specific settings'
        },
        category: {
          bsonType: 'string',
          description: 'Setting category'
        },
        created_at: {
          bsonType: 'date',
          description: 'Creation timestamp'
        },
        updated_at: {
          bsonType: 'date',
          description: 'Last update timestamp'
        }
      }
    }
  }
});

// Create indexes for system settings
db.system_settings.createIndex({ key: 1 }, { unique: true });
db.system_settings.createIndex({ tenant_id: 1 });
db.system_settings.createIndex({ category: 1 });

// Create file storage collection
db.createCollection('files', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['filename', 'tenant_id', 'file_path'],
      properties: {
        filename: {
          bsonType: 'string',
          description: 'Original filename'
        },
        tenant_id: {
          bsonType: 'string',
          description: 'Tenant ID who owns the file'
        },
        file_path: {
          bsonType: 'string',
          description: 'Storage path of the file'
        },
        file_size: {
          bsonType: 'number',
          description: 'File size in bytes'
        },
        mime_type: {
          bsonType: 'string',
          description: 'MIME type of the file'
        },
        metadata: {
          bsonType: 'object',
          description: 'Additional file metadata'
        },
        uploaded_by: {
          bsonType: 'string',
          description: 'User ID who uploaded the file'
        },
        created_at: {
          bsonType: 'date',
          description: 'Upload timestamp'
        }
      }
    }
  }
});

// Create indexes for files
db.files.createIndex({ tenant_id: 1 });
db.files.createIndex({ uploaded_by: 1 });
db.files.createIndex({ mime_type: 1 });
db.files.createIndex({ created_at: -1 });

// Create logs collection for application logs
db.createCollection('application_logs', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['level', 'message', 'timestamp'],
      properties: {
        level: {
          bsonType: 'string',
          enum: ['debug', 'info', 'warn', 'error', 'fatal'],
          description: 'Log level'
        },
        message: {
          bsonType: 'string',
          description: 'Log message'
        },
        tenant_id: {
          bsonType: 'string',
          description: 'Tenant ID if applicable'
        },
        user_id: {
          bsonType: 'string',
          description: 'User ID if applicable'
        },
        context: {
          bsonType: 'object',
          description: 'Additional log context'
        },
        timestamp: {
          bsonType: 'date',
          description: 'Log timestamp'
        }
      }
    }
  }
});

// Create indexes for logs
db.application_logs.createIndex({ timestamp: -1 });
db.application_logs.createIndex({ level: 1 });
db.application_logs.createIndex({ tenant_id: 1 });
db.application_logs.createIndex({ user_id: 1 });

// Insert initial system settings
db.system_settings.insertMany([
  {
    key: 'app_name',
    value: 'Zplus SaaS Base',
    category: 'general',
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    key: 'app_version',
    value: '1.0.0',
    category: 'general',
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    key: 'max_file_size',
    value: 10485760, // 10MB
    category: 'files',
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    key: 'allowed_file_types',
    value: ['jpg', 'jpeg', 'png', 'gif', 'pdf', 'doc', 'docx', 'xls', 'xlsx'],
    category: 'files',
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    key: 'email_verification_required',
    value: true,
    category: 'security',
    created_at: new Date(),
    updated_at: new Date()
  }
]);

// Create sample tenant databases (these would normally be created dynamically)
var sampleTenants = ['demo', 'acme', 'startup', 'enterprise'];

sampleTenants.forEach(function(tenantId) {
  var tenantDb = db.getSiblingDB('tenant_' + tenantId);
  
  // Create tenant-specific collections
  tenantDb.createCollection('documents', {
    validator: {
      $jsonSchema: {
        bsonType: 'object',
        required: ['title', 'content', 'created_by'],
        properties: {
          title: {
            bsonType: 'string',
            description: 'Document title'
          },
          content: {
            bsonType: 'string',
            description: 'Document content'
          },
          created_by: {
            bsonType: 'string',
            description: 'User ID who created the document'
          },
          tags: {
            bsonType: 'array',
            items: { bsonType: 'string' },
            description: 'Document tags'
          },
          created_at: {
            bsonType: 'date',
            description: 'Creation timestamp'
          },
          updated_at: {
            bsonType: 'date',
            description: 'Last update timestamp'
          }
        }
      }
    }
  });
  
  // Create indexes for tenant documents
  tenantDb.documents.createIndex({ created_by: 1 });
  tenantDb.documents.createIndex({ tags: 1 });
  tenantDb.documents.createIndex({ created_at: -1 });
  tenantDb.documents.createIndex({ title: 'text', content: 'text' });
  
  // Insert sample documents
  tenantDb.documents.insertMany([
    {
      title: 'Welcome to ' + tenantId.toUpperCase(),
      content: 'This is your first document in the ' + tenantId + ' tenant.',
      created_by: 'sample_user_id',
      tags: ['welcome', 'getting-started'],
      created_at: new Date(),
      updated_at: new Date()
    },
    {
      title: 'Sample Document',
      content: 'This is a sample document for testing purposes.',
      created_by: 'sample_user_id',
      tags: ['sample', 'test'],
      created_at: new Date(),
      updated_at: new Date()
    }
  ]);
});

print('MongoDB initialization completed successfully');
