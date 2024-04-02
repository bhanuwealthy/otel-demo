// mongo-init.js

var otel_db = db.getSiblingDB('demoDb');
db.createUser({
  user: 'otel',
  pwd: 'otel',
  roles: [
    {
      role: 'readWrite',
      db: 'demoDb'
    }
  ]
});

otel_db.createCollection('otel_tenants');


// Check if 'Tenant 1' already exists in the 'tenants' collection
var existingTenant = otel_db.otel_tenants.findOne({ client_id: 'tonystark' });

// If 'Tenant 1' does not exist, insert it
if (!existingTenant) {
  otel_db.otel_tenants.insert({
    client_id: 'tonystark',
    secret_key: 'jarvis',
    client_name: 'Local' });
}
